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
