// Package site holds the site model: pages, output-path mapping and the
// navigation tree.
package site

import (
	"fmt"
	"path"
	"sort"
	"strings"
)

type Page struct {
	SrcRel  string // slash path relative to input root, e.g. "docs/setup.md"
	OutRel  string // slash path relative to output root, e.g. "docs/setup/index.html"
	Title   string // first h1, or file name without extension
	IsIndex bool   // true for README.md pages
	Source  []byte
}

// Depth is the number of directories between the page's output file and the
// site root; used to build "../" prefixes.
func (p *Page) Depth() int {
	dir := path.Dir(p.OutRel)
	if dir == "." {
		return 0
	}
	return strings.Count(dir, "/") + 1
}

// RelRoot returns the "../" chain from the page back to the site root.
func (p *Page) RelRoot() string {
	return strings.Repeat("../", p.Depth())
}

type Site struct {
	Title    string
	Pages    []*Page
	bySrc    map[string]*Page // lowercased SrcRel -> page
	Root     *NavNode
	Warnings []string
}

type NavNode struct {
	Label    string
	Page     *Page // nil for directories without a README
	Dir      string
	Children []*NavNode
}

func (s *Site) Warnf(format string, args ...any) {
	s.Warnings = append(s.Warnings, fmt.Sprintf(format, args...))
}

// LookupSrc resolves an input-relative markdown path to its page.
func (s *Site) LookupSrc(srcRel string) *Page {
	return s.bySrc[strings.ToLower(path.Clean(srcRel))]
}

func isReadme(name string) bool {
	return strings.EqualFold(name, "README.md")
}

// New maps the scanned markdown files to pages and builds the nav tree.
// Sources must be loaded by the caller afterwards; here only paths are laid
// out so collisions can be resolved globally.
func New(files []string) *Site {
	s := &Site{bySrc: map[string]*Page{}}

	readmeDirs := map[string]bool{} // dirs that have a README.md
	for _, f := range files {
		if isReadme(path.Base(f)) {
			readmeDirs[path.Dir(f)] = true
		}
	}

	for _, f := range files {
		dir, base := path.Dir(f), path.Base(f)
		stem := strings.TrimSuffix(base, path.Ext(base))
		p := &Page{SrcRel: f}
		switch {
		case isReadme(base):
			p.IsIndex = true
			if dir == "." {
				p.OutRel = "index.html"
			} else {
				p.OutRel = dir + "/index.html"
			}
		case readmeDirs[path.Join(dir, stem)]:
			// foo.md next to foo/README.md would collide on foo/index.html:
			// the README wins, this page falls back to flat foo.html.
			p.OutRel = strings.TrimSuffix(f, path.Ext(f)) + ".html"
			s.Warnf("%s collides with %s; writing flat %s", f, path.Join(dir, stem, "README.md"), p.OutRel)
		default:
			p.OutRel = strings.TrimSuffix(f, path.Ext(f)) + "/index.html"
		}
		s.Pages = append(s.Pages, p)
		s.bySrc[strings.ToLower(f)] = p
	}
	return s
}

// BuildNav constructs the sidebar tree from the pages. Called after titles
// have been extracted. The root README is excluded (it is linked via the
// site title) unless it is the only page.
func (s *Site) BuildNav() {
	s.Root = &NavNode{Dir: "."}
	dirNodes := map[string]*NavNode{".": s.Root}

	var nodeFor func(dir string) *NavNode
	nodeFor = func(dir string) *NavNode {
		if n, ok := dirNodes[dir]; ok {
			return n
		}
		parent := nodeFor(path.Dir(dir))
		n := &NavNode{Label: path.Base(dir), Dir: dir}
		parent.Children = append(parent.Children, n)
		dirNodes[dir] = n
		return n
	}

	for _, p := range s.Pages {
		if p.SrcRel == "README.md" || strings.EqualFold(p.SrcRel, "readme.md") {
			continue // root README is the site-title link
		}
		dir := path.Dir(p.SrcRel)
		if p.IsIndex {
			n := nodeFor(dir)
			n.Page = p
			n.Label = p.Title
		} else {
			parent := nodeFor(dir)
			parent.Children = append(parent.Children, &NavNode{Label: p.Title, Page: p, Dir: dir})
		}
	}
	sortNav(s.Root)
}

func sortNav(n *NavNode) {
	sort.SliceStable(n.Children, func(i, j int) bool {
		a, b := n.Children[i], n.Children[j]
		// Leaf pages before subdirectories, then alphabetical.
		if (len(a.Children) == 0) != (len(b.Children) == 0) {
			return len(a.Children) == 0
		}
		return strings.ToLower(a.Label) < strings.ToLower(b.Label)
	})
	for _, c := range n.Children {
		sortNav(c)
	}
}
