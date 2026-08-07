package scan

import (
	"reflect"
	"testing"
	"testing/fstest"
)

func TestSourcesIncludesTabular(t *testing.T) {
	fsys := fstest.MapFS{
		"README.md": {Data: []byte("x")},
		"a.csv":     {Data: []byte("x")},
		"b.tsv":     {Data: []byte("x")},
		"c.txt":     {Data: []byte("x")},
	}
	got, err := Sources(fsys, nil)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"README.md", "a.csv", "b.tsv"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Sources() = %v, want %v", got, want)
	}
}

func TestMarkdown(t *testing.T) {
	fsys := fstest.MapFS{
		"README.md":                {Data: []byte("x")},
		"docs/setup.md":            {Data: []byte("x")},
		"docs/notes.txt":           {Data: []byte("x")},
		"node_modules/a/b.md":      {Data: []byte("x")},
		"vendor/c.md":              {Data: []byte("x")},
		".git/d.md":                {Data: []byte("x")},
		".hidden/e.md":             {Data: []byte("x")},
		"docs/.secret.md":          {Data: []byte("x")},
		"drafts/wip.md":            {Data: []byte("x")},
		"CHANGELOG.md":             {Data: []byte("x")},
		"deep/nested/dir/page.md":  {Data: []byte("x")},
		"deep/nested/skip/page.md": {Data: []byte("x")},
	}
	got, err := Markdown(fsys, []string{"drafts", "CHANGELOG.md", "deep/nested/skip"})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"README.md", "deep/nested/dir/page.md", "docs/setup.md"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Markdown() = %v, want %v", got, want)
	}
}

func TestMarkdownGlobExclude(t *testing.T) {
	fsys := fstest.MapFS{
		"a.md":          {Data: []byte("x")},
		"a_test.md":     {Data: []byte("x")},
		"sub/b_test.md": {Data: []byte("x")},
	}
	got, err := Markdown(fsys, []string{"*_test.md"})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"a.md"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Markdown() = %v, want %v", got, want)
	}
}
