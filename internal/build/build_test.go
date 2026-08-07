package build

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunFixture(t *testing.T) {
	out := filepath.Join(t.TempDir(), "public")
	var log bytes.Buffer
	err := Run(Config{InputDir: "testdata/repo", OutputDir: out, Log: &log})
	if err != nil {
		t.Fatalf("Run: %v\n%s", err, log.String())
	}

	for _, f := range []string{
		"index.html",
		"guide/index.html",
		"docs/index.html",
		"docs/setup/index.html",
		"data/index.html",
		"metrics/index.html",
		"assets/style.css",
		"assets/app.js",
		"assets/search-index.js",
		"img/logo.png",
		MarkerName,
	} {
		if _, err := os.Stat(filepath.Join(out, f)); err != nil {
			t.Errorf("missing output file %s", f)
		}
	}
	// Excluded sources must not leak into the output.
	if _, err := os.Stat(filepath.Join(out, "node_modules")); err == nil {
		t.Error("node_modules leaked into output")
	}

	index := readOut(t, out, "index.html")
	for _, want := range []string{
		"<title>サンプルプロジェクト</title>",
		`href="guide/index.html"`,
		`href="docs/setup/index.html"`,
		`href="docs/index.html"`, // directory link resolved to its README
		`<img src="img/logo.png"`,
		`class="page-updated"`,
		`id="view-controls"`,
	} {
		if !strings.Contains(index, want) {
			t.Errorf("index.html missing %s", want)
		}
	}

	setup := readOut(t, out, "docs/setup/index.html")
	for _, want := range []string{
		`href="../../guide/index.html#usage"`,
		`href="../../assets/style.css"`,
		"chroma", // highlighted code block
	} {
		if !strings.Contains(setup, want) {
			t.Errorf("docs/setup/index.html missing %s", want)
		}
	}

	csvPage := readOut(t, out, "data/index.html")
	for _, want := range []string{
		"<h1>data</h1>",
		`<table class="tabular-table">`,
		"<th>name</th>",
		"<td>Alice</td>",
	} {
		if !strings.Contains(csvPage, want) {
			t.Errorf("data/index.html missing %s", want)
		}
	}

	tsvPage := readOut(t, out, "metrics/index.html")
	if !strings.Contains(tsvPage, "<th>metric</th>") || !strings.Contains(tsvPage, "<td>145</td>") {
		t.Errorf("metrics/index.html missing tsv table content")
	}

	searchJS := readOut(t, out, "assets/search-index.js")
	for _, want := range []string{"サンプルプロジェクト", "日本語の検索テスト用テキスト", "guide/index.html", "Alice", "users"} {
		if !strings.Contains(searchJS, want) {
			t.Errorf("search-index.js missing %s", want)
		}
	}
}

func TestRunRefusesUnmarkedOutputDir(t *testing.T) {
	out := t.TempDir()
	if err := os.WriteFile(filepath.Join(out, "precious.txt"), []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}
	var log bytes.Buffer
	err := Run(Config{InputDir: "testdata/repo", OutputDir: out, Log: &log})
	if err == nil || !strings.Contains(err.Error(), MarkerName) {
		t.Fatalf("want marker refusal error, got %v", err)
	}
	if _, err := os.Stat(filepath.Join(out, "precious.txt")); err != nil {
		t.Error("existing file was destroyed")
	}
}

func TestRunRebuildsMarkedDir(t *testing.T) {
	out := filepath.Join(t.TempDir(), "public")
	var log bytes.Buffer
	cfg := Config{InputDir: "testdata/repo", OutputDir: out, Log: &log}
	if err := Run(cfg); err != nil {
		t.Fatal(err)
	}
	// Second build over a marked directory succeeds and replaces content.
	if err := Run(cfg); err != nil {
		t.Fatalf("rebuild over marked dir: %v", err)
	}
}

func readOut(t *testing.T, out, rel string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(out, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
