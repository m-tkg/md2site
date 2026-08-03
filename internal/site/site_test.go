package site

import "testing"

func TestOutputMapping(t *testing.T) {
	s := New([]string{
		"README.md",
		"guide.md",
		"docs/README.md",
		"docs/setup.md",
		"docs.md", // collides with docs/README.md
	})
	want := map[string]string{
		"README.md":      "index.html",
		"guide.md":       "guide/index.html",
		"docs/README.md": "docs/index.html",
		"docs/setup.md":  "docs/setup/index.html",
		"docs.md":        "docs.html",
	}
	for src, out := range want {
		p := s.LookupSrc(src)
		if p == nil {
			t.Fatalf("page %s not found", src)
		}
		if p.OutRel != out {
			t.Errorf("%s -> %s, want %s", src, p.OutRel, out)
		}
	}
	if len(s.Warnings) != 1 {
		t.Errorf("want 1 collision warning, got %v", s.Warnings)
	}
}

func TestDepthAndRelRoot(t *testing.T) {
	cases := []struct {
		out   string
		depth int
		rel   string
	}{
		{"index.html", 0, ""},
		{"guide/index.html", 1, "../"},
		{"docs/setup/index.html", 2, "../../"},
	}
	for _, c := range cases {
		p := &Page{OutRel: c.out}
		if p.Depth() != c.depth || p.RelRoot() != c.rel {
			t.Errorf("%s: depth=%d rel=%q, want %d %q", c.out, p.Depth(), p.RelRoot(), c.depth, c.rel)
		}
	}
}

func TestBuildNav(t *testing.T) {
	s := New([]string{"README.md", "guide.md", "docs/README.md", "docs/setup.md"})
	for _, p := range s.Pages {
		p.Title = p.SrcRel // stand-in titles
	}
	s.BuildNav()

	if len(s.Root.Children) != 2 {
		t.Fatalf("root children = %d, want 2 (guide + docs)", len(s.Root.Children))
	}
	// Leaf pages sort before directories.
	if s.Root.Children[0].Page == nil || s.Root.Children[0].Page.SrcRel != "guide.md" {
		t.Errorf("first child should be guide.md leaf, got %+v", s.Root.Children[0])
	}
	docs := s.Root.Children[1]
	if docs.Page == nil || docs.Page.SrcRel != "docs/README.md" {
		t.Errorf("docs node should link its README, got %+v", docs.Page)
	}
	if len(docs.Children) != 1 || docs.Children[0].Page.SrcRel != "docs/setup.md" {
		t.Errorf("docs children wrong: %+v", docs.Children)
	}
}
