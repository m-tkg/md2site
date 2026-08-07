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

func mustParse(t *testing.T, p *site.Page) *Doc {
	t.Helper()
	d, err := Parse(p)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func TestTitleAndPlainText(t *testing.T) {
	s := newSite(t, map[string]string{
		"a.md": "# 見出しタイトル\n\n本文テキスト。\n\n```sh\necho hi\n```\n",
	})
	d := mustParse(t, s.LookupSrc("a.md"))
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
		"docs/setup.md": "[g](../guide.md#usage) [top](../README.md) [csv](../data.csv) [ext](https://example.com/a.md) [missing](nope.md) ![i](img/shot.png)\n\n<img src=\"img/raw.png\" alt=\"raw html\">\n",
		"data.csv":      "a,b\n1,2\n",
	})
	p := s.LookupSrc("docs/setup.md")
	d := mustParse(t, p)
	body, assets, err := d.Body(s)
	if err != nil {
		t.Fatal(err)
	}
	html := string(body)
	for _, want := range []string{
		`href="../../guide/index.html#usage"`,
		`href="../../index.html"`,
		`href="../../data/index.html"`,
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

func TestTabularCSV(t *testing.T) {
	s := newSite(t, map[string]string{
		"team.csv": "name,role\nAlice,Admin\nBob,Editor\n",
	})
	p := s.LookupSrc("team.csv")
	p.Title = "team"
	d := mustParse(t, p)
	if len(d.Outline()) != 0 {
		t.Fatalf("tabular outline should be empty")
	}
	body, _, err := d.Body(s)
	if err != nil {
		t.Fatal(err)
	}
	html := string(body)
	for _, want := range []string{
		"<h1>team</h1>",
		"<th>name</th>",
		"<th>role</th>",
		"<td>Alice</td>",
		"<td>Admin</td>",
		`<table class="tabular-table">`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("body missing %q\n%s", want, html)
		}
	}
	if txt := d.PlainText(); !strings.Contains(txt, "Alice Admin Bob Editor") {
		t.Errorf("PlainText() = %q", txt)
	}
}

func TestTabularTSV(t *testing.T) {
	s := newSite(t, map[string]string{
		"log.tsv": "date\tvalue\n2026-01-01\t42\n",
	})
	p := s.LookupSrc("log.tsv")
	p.Title = "log"
	d := mustParse(t, p)
	body, _, err := d.Body(s)
	if err != nil {
		t.Fatal(err)
	}
	html := string(body)
	if !strings.Contains(html, "<th>date</th>") || !strings.Contains(html, "<td>42</td>") {
		t.Errorf("unexpected tsv body: %s", html)
	}
}

func TestOutline(t *testing.T) {
	s := newSite(t, map[string]string{
		"a.md": "# セットアップ手順\n\n## インストール\n\n### Usage\n\n## インストール\n\n本文\n\n##### deep\n",
	})
	d := mustParse(t, s.LookupSrc("a.md"))
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
		d := mustParse(t, p)
		p.Title = d.Title()
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
