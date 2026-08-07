package render_test

import (
	"strings"
	"testing"

	"github.com/m-tkg/md2site/internal/render"
	"github.com/m-tkg/md2site/internal/site"
)

func TestFrontMatterTitle(t *testing.T) {
	s := site.New([]string{"README.md"})
	p := s.LookupSrc("README.md")
	p.Source = []byte(`---
title: "サービスカタログ"
description: "CSVを正本としてGitHub Pagesに表示するサービスカタログ。"
last_reviewed: 2026-05-28
---

# 別の見出し

本文
`)
	d, err := render.Parse(p)
	if err != nil {
		t.Fatal(err)
	}
	if got := d.Title(); got != "サービスカタログ" {
		t.Errorf("Title() = %q, want サービスカタログ", got)
	}
	body, _, err := d.Body(s)
	if err != nil {
		t.Fatal(err)
	}
	html := string(body)
	if strings.Contains(html, "last_reviewed:") || strings.Contains(html, "description:") {
		t.Errorf("front matter leaked into body: %s", html)
	}
	if !strings.Contains(html, "別の見出し") {
		t.Errorf("body missing markdown heading: %s", html)
	}
}

func TestFrontMatterWithoutHeading(t *testing.T) {
	s := site.New([]string{"README.md"})
	p := s.LookupSrc("README.md")
	p.Source = []byte(`---
title: "サービスカタログ"
description: "CSVを正本としてGitHub Pagesに表示するサービスカタログ。"
last_reviewed: 2026-05-28
---

本文のみ
`)
	d, err := render.Parse(p)
	if err != nil {
		t.Fatal(err)
	}
	if got := d.Title(); got != "サービスカタログ" {
		t.Errorf("Title() = %q", got)
	}
}
