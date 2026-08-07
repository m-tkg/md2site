// Package render converts markdown to HTML: goldmark parsing, relative-link
// rewriting over the rendered HTML, plain-text extraction for search, and
// sidebar navigation rendering.
package render

import (
	"bytes"
	"fmt"
	"html"
	"html/template"
	"path"
	"regexp"
	"strings"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	goldmarkhtml "github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"

	"github.com/m-tkg/md2site/internal/site"
)

var md = goldmark.New(
	goldmark.WithExtensions(
		extension.GFM,
		highlighting.NewHighlighting(
			highlighting.WithFormatOptions(chromahtml.WithClasses(true)),
		),
	),
	goldmark.WithParserOptions(parser.WithAutoHeadingID()),
	goldmark.WithRendererOptions(goldmarkhtml.WithUnsafe()),
)

// Doc is a parsed page: markdown AST or tabular rows.
type Doc struct {
	page    *site.Page
	root    ast.Node
	table   [][]string
	fmTitle string
	body    []byte // markdown body with front matter removed
}

// Parse parses the page source. Call before title extraction or rendering.
func Parse(p *site.Page) (*Doc, error) {
	if p.IsTabular() {
		tsv := strings.EqualFold(path.Ext(p.SrcRel), ".tsv")
		rows, err := parseTabular(p.Source, tsv)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", p.SrcRel, err)
		}
		return &Doc{page: p, table: rows}, nil
	}
	ctx := parser.NewContext(parser.WithIDs(newIDs()))
	fm, body, stripped := stripFrontMatter(p.Source)
	src := p.Source
	if stripped {
		src = body
	}
	root := md.Parser().Parse(text.NewReader(src), parser.WithContext(ctx))
	return &Doc{page: p, root: root, fmTitle: strings.TrimSpace(fm.Title), body: src}, nil
}

// Heading is one outline entry of a page.
type Heading struct {
	Level int
	Text  string
	ID    string
}

// outlineMaxLevel bounds the outline depth; deeper headings stay in the
// body but out of the outline column.
const outlineMaxLevel = 4

// Outline lists the page's headings (levels 1..outlineMaxLevel) in order.
func (d *Doc) Outline() []Heading {
	if d.table != nil {
		return nil
	}
	var hs []Heading
	ast.Walk(d.root, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if h, ok := n.(*ast.Heading); ok && entering && h.Level <= outlineMaxLevel {
			id := ""
			if v, ok := h.AttributeString("id"); ok {
				if b, ok := v.([]byte); ok {
					id = string(b)
				}
			}
			hs = append(hs, Heading{Level: h.Level, Text: strings.TrimSpace(nodeText(h, d.mdSource())), ID: id})
		}
		return ast.WalkContinue, nil
	})
	return hs
}

// OutlineHTML renders the outline as nested lists of in-page anchor links.
// Returns "" when the page has at most one heading (an outline of just the
// title carries no information).
func OutlineHTML(hs []Heading) template.HTML {
	if len(hs) <= 1 {
		return ""
	}
	var b strings.Builder
	var stack []int // heading levels of the open lists
	for _, h := range hs {
		switch {
		case len(stack) == 0 || h.Level > stack[len(stack)-1]:
			// One list deeper, even when heading levels skip (h2 -> h4).
			b.WriteString("<ul>")
			stack = append(stack, h.Level)
		default:
			for len(stack) > 1 && h.Level < stack[len(stack)-1] {
				b.WriteString("</li></ul>")
				stack = stack[:len(stack)-1]
			}
			b.WriteString("</li>")
		}
		fmt.Fprintf(&b, "<li><a href=\"#%s\">%s</a>", h.ID, html.EscapeString(h.Text))
	}
	for range stack {
		b.WriteString("</li></ul>")
	}
	return template.HTML(b.String())
}

// Title returns the text of the first level-1 heading, or "".
func (d *Doc) Title() string {
	if d.table != nil {
		return ""
	}
	if d.fmTitle != "" {
		return d.fmTitle
	}
	title := ""
	src := d.mdSource()
	ast.Walk(d.root, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if h, ok := n.(*ast.Heading); ok && entering && h.Level == 1 {
			title = nodeText(h, src)
			return ast.WalkStop, nil
		}
		return ast.WalkContinue, nil
	})
	return strings.TrimSpace(title)
}

// PlainText extracts the searchable text of the page (body and code).
func (d *Doc) PlainText() string {
	if d.table != nil {
		return tablePlainText(d.table)
	}
	var b strings.Builder
	src := d.mdSource()
	ast.Walk(d.root, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch t := n.(type) {
		case *ast.Text:
			b.Write(t.Segment.Value(src))
			b.WriteByte(' ')
		case *ast.String:
			b.Write(t.Value)
			b.WriteByte(' ')
		case *ast.FencedCodeBlock, *ast.CodeBlock:
			lines := n.Lines()
			for i := 0; i < lines.Len(); i++ {
				seg := lines.At(i)
				b.Write(seg.Value(src))
			}
			b.WriteByte(' ')
		}
		return ast.WalkContinue, nil
	})
	return strings.Join(strings.Fields(b.String()), " ")
}

var attrRe = regexp.MustCompile(`(src|href)=(["'])([^"']*)(["'])`)

// Body renders the AST and rewrites relative src/href destinations in the
// resulting HTML. Operating on the rendered HTML covers both markdown links
// and raw HTML fragments (e.g. <img> tags common in READMEs) with one code
// path. It returns the input-relative paths of referenced local assets.
func (d *Doc) Body(s *site.Site) (template.HTML, []string, error) {
	if d.table != nil {
		title := d.page.Title
		if title == "" {
			base := path.Base(d.page.SrcRel)
			title = strings.TrimSuffix(base, path.Ext(base))
		}
		body, err := TableHTML(title, d.table)
		return body, nil, err
	}

	var buf bytes.Buffer
	if err := md.Renderer().Render(&buf, d.mdSource(), d.root); err != nil {
		return "", nil, fmt.Errorf("render %s: %w", d.page.SrcRel, err)
	}

	var assets []string
	relRoot := d.page.RelRoot()
	srcDir := path.Dir(d.page.SrcRel)

	rewrite := func(dest string, isAsset bool) (string, bool) {
		if dest == "" || strings.HasPrefix(dest, "#") || strings.HasPrefix(dest, "/") ||
			strings.Contains(dest, "://") || strings.HasPrefix(dest, "mailto:") ||
			strings.HasPrefix(dest, "data:") {
			return "", false
		}
		target, frag := dest, ""
		if i := strings.IndexByte(dest, '#'); i >= 0 {
			target, frag = dest[:i], dest[i:]
		}
		if target == "" {
			return "", false
		}
		resolved := path.Join(srcDir, target)
		if resolved == ".." || strings.HasPrefix(resolved, "../") {
			s.Warnf("%s: link %q escapes the input directory; left as-is", d.page.SrcRel, dest)
			return "", false
		}
		if isAsset {
			assets = append(assets, resolved)
			return relRoot + resolved, true
		}
		if isPageExt(path.Ext(target)) {
			if p := s.LookupSrc(resolved); p != nil {
				return relRoot + p.OutRel + frag, true
			}
			s.Warnf("%s: link to missing file %q; left as-is", d.page.SrcRel, dest)
			return "", false
		}
		// Directory link -> its README page, if any.
		if p := s.LookupSrc(path.Join(resolved, "README.md")); p != nil {
			return relRoot + p.OutRel + frag, true
		}
		return "", false
	}

	out := attrRe.ReplaceAllStringFunc(buf.String(), func(m string) string {
		sub := attrRe.FindStringSubmatch(m)
		attr, quote, val := sub[1], sub[2], sub[3]
		// URLs in rendered HTML are entity-escaped; unescape before
		// resolving and re-escape on output.
		if nd, ok := rewrite(html.UnescapeString(val), attr == "src"); ok {
			return attr + "=" + quote + html.EscapeString(nd) + quote
		}
		return m
	})
	return template.HTML(out), assets, nil
}

func isPageExt(ext string) bool {
	switch strings.ToLower(ext) {
	case ".md", ".csv", ".tsv":
		return true
	default:
		return false
	}
}

func nodeText(n ast.Node, source []byte) string {
	var b strings.Builder
	ast.Walk(n, func(c ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch t := c.(type) {
		case *ast.Text:
			b.Write(t.Segment.Value(source))
		case *ast.String:
			b.Write(t.Value)
		}
		return ast.WalkContinue, nil
	})
	return b.String()
}

// NavHTML renders the sidebar tree for one page (links are page-relative).
func NavHTML(s *site.Site, cur *site.Page) template.HTML {
	var b strings.Builder
	relRoot := cur.RelRoot()
	renderNodes(&b, s.Root.Children, cur, relRoot)
	return template.HTML(b.String())
}

func renderNodes(b *strings.Builder, nodes []*site.NavNode, cur *site.Page, relRoot string) {
	if len(nodes) == 0 {
		return
	}
	b.WriteString("<ul>")
	for _, n := range nodes {
		b.WriteString("<li>")
		if len(n.Children) > 0 {
			open := ""
			if underDir(cur, n.Dir) {
				open = " open"
			}
			fmt.Fprintf(b, "<details%s><summary>", open)
			writeNavLabel(b, n, cur, relRoot)
			b.WriteString("</summary>")
			renderNodes(b, n.Children, cur, relRoot)
			b.WriteString("</details>")
		} else {
			writeNavLabel(b, n, cur, relRoot)
		}
		b.WriteString("</li>")
	}
	b.WriteString("</ul>")
}

func writeNavLabel(b *strings.Builder, n *site.NavNode, cur *site.Page, relRoot string) {
	label := html.EscapeString(n.Label)
	if n.Page == nil {
		fmt.Fprintf(b, "<span class=\"nav-dir\">%s</span>", label)
		return
	}
	class := "nav-link"
	if n.Page == cur {
		class += " current"
	}
	fmt.Fprintf(b, "<a class=%q href=%q>%s</a>", class, relRoot+n.Page.OutRel, label)
}

func underDir(p *site.Page, dir string) bool {
	if dir == "." {
		return true
	}
	return strings.HasPrefix(p.SrcRel, dir+"/")
}
