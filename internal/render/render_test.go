package render

import (
	"strings"
	"testing"

	"github.com/m-tkg/md2site/internal/site"
)

func newSite(t *testing.T, sources map[string]string) *site.Site {
	t.Helper()
	files := make([]string, 0, len(sources))
	for f := range sources {
		files = append(files, f)
	}
	s := site.New(files)
	for f, src := range sources {
		s.LookupSrc(f).Source = []byte(src)
	}
	return s
}

func TestTitleAndPlainText(t *testing.T) {
	s := newSite(t, map[string]string{
		"a.md": "# 見出しタイトル\n\n本文テキスト。\n\n```sh\necho hi\n```\n",
	})
	d := Parse(s.LookupSrc("a.md"))
	if got := d.Title(); got != "見出しタイトル" {
		t.Errorf("Title() = %q", got)
	}
	txt := d.PlainText()
	for _, want := range []string{"見出しタイトル", "本文テキスト。", "echo hi"} {
		if !strings.Contains(txt, want) {
			t.Errorf("PlainText() missing %q: %q", want, txt)
		}
	}
}

func TestRewriteLinks(t *testing.T) {
	s := newSite(t, map[string]string{
		"README.md":     "# top\n",
		"guide.md":      "# guide\n",
		"docs/setup.md": "[g](../guide.md#usage) [top](../README.md) [ext](https://example.com/a.md) [missing](nope.md) ![i](img/shot.png)\n\n<img src=\"img/raw.png\" alt=\"raw html\">\n",
	})
	p := s.LookupSrc("docs/setup.md")
	d := Parse(p)
	body, assets, err := d.Body(s)
	if err != nil {
		t.Fatal(err)
	}
	html := string(body)
	for _, want := range []string{
		`href="../../guide/index.html#usage"`,
		`href="../../index.html"`,
		`href="https://example.com/a.md"`, // external untouched
		`href="nope.md"`,                  // missing untouched
		`src="../../docs/img/shot.png"`,
		`src="../../docs/img/raw.png"`, // raw HTML <img> also rewritten
	} {
		if !strings.Contains(html, want) {
			t.Errorf("body missing %s\n%s", want, html)
		}
	}
	if len(assets) != 2 || assets[0] != "docs/img/shot.png" || assets[1] != "docs/img/raw.png" {
		t.Errorf("assets = %v", assets)
	}
	if len(s.Warnings) != 1 || !strings.Contains(s.Warnings[0], "nope.md") {
		t.Errorf("warnings = %v", s.Warnings)
	}
}

func TestOutline(t *testing.T) {
	s := newSite(t, map[string]string{
		"a.md": "# セットアップ手順\n\n## インストール\n\n### Usage\n\n## インストール\n\n本文\n\n##### deep\n",
	})
	d := Parse(s.LookupSrc("a.md"))
	hs := d.Outline()
	want := []Heading{
		{1, "セットアップ手順", "セットアップ手順"},
		{2, "インストール", "インストール"},
		{3, "Usage", "usage"},
		{2, "インストール", "インストール-1"}, // duplicates uniquified
	}
	if len(hs) != len(want) { // h5 excluded by outlineMaxLevel
		t.Fatalf("Outline() = %+v", hs)
	}
	for i, w := range want {
		if hs[i] != w {
			t.Errorf("Outline()[%d] = %+v, want %+v", i, hs[i], w)
		}
	}
	out := string(OutlineHTML(hs))
	if !strings.Contains(out, `<a href="#セットアップ手順">セットアップ手順</a>`) ||
		!strings.Contains(out, `<a href="#インストール-1">`) {
		t.Errorf("OutlineHTML = %s", out)
	}
	if strings.Count(out, "<ul>") != strings.Count(out, "</ul>") ||
		strings.Count(out, "<li>") != strings.Count(out, "</li>") {
		t.Errorf("unbalanced outline HTML: %s", out)
	}
}

func TestOutlineHTMLStartsDeep(t *testing.T) {
	// First heading deeper than a later one must not over-close lists.
	out := string(OutlineHTML([]Heading{{2, "a", "a"}, {4, "b", "b"}, {1, "c", "c"}}))
	if strings.Count(out, "<ul>") != strings.Count(out, "</ul>") ||
		strings.Count(out, "<li>") != strings.Count(out, "</li>") {
		t.Errorf("unbalanced outline HTML: %s", out)
	}
}

func TestOutlineHTMLSingleHeading(t *testing.T) {
	if out := OutlineHTML([]Heading{{1, "only", "only"}}); out != "" {
		t.Errorf("single-heading outline should be empty, got %s", out)
	}
}

func TestNavHTMLCurrentPage(t *testing.T) {
	s := newSite(t, map[string]string{
		"README.md": "# top\n",
		"guide.md":  "# ガイド\n",
	})
	for _, p := range s.Pages {
		p.Title = Parse(p).Title()
	}
	s.BuildNav()
	nav := string(NavHTML(s, s.LookupSrc("guide.md")))
	if !strings.Contains(nav, `class="nav-link current"`) {
		t.Errorf("current page not highlighted: %s", nav)
	}
	if !strings.Contains(nav, `href="../guide/index.html"`) {
		t.Errorf("nav link not page-relative: %s", nav)
	}
}
