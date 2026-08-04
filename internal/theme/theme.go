// Package theme embeds the built-in minimal theme and writes its static
// assets (including generated chroma stylesheets) to the output directory.
package theme

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"os"
	"path/filepath"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/styles"
)

//go:embed layout.html
var layoutSrc string

//go:embed style.css
var styleCSS []byte

//go:embed app.js
var appJS []byte

var Layout = template.Must(template.New("layout").Parse(layoutSrc))

// PageData is the input to the layout template.
type PageData struct {
	SiteTitle string
	Title     string
	RelRoot   string
	Nav       template.HTML
	Outline   template.HTML // empty when the page has at most one heading
	Content   template.HTML
}

// WriteAssets writes style.css (with chroma highlight styles appended) and
// app.js under <outDir>/assets/.
func WriteAssets(outDir string) error {
	assetsDir := filepath.Join(outDir, "assets")
	if err := os.MkdirAll(assetsDir, 0o755); err != nil {
		return err
	}
	css, err := fullCSS()
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(assetsDir, "style.css"), css, 0o644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(assetsDir, "app.js"), appJS, 0o644)
}

func fullCSS() ([]byte, error) {
	var buf bytes.Buffer
	buf.Write(styleCSS)
	formatter := chromahtml.New(chromahtml.WithClasses(true))

	buf.WriteString("\n/* chroma (light) */\n")
	if err := formatter.WriteCSS(&buf, styles.Get("github")); err != nil {
		return nil, fmt.Errorf("chroma css: %w", err)
	}
	buf.WriteString("\n@media (prefers-color-scheme: dark) {\n")
	var dark bytes.Buffer
	if err := formatter.WriteCSS(&dark, styles.Get("github-dark")); err != nil {
		return nil, fmt.Errorf("chroma dark css: %w", err)
	}
	buf.Write(dark.Bytes())
	buf.WriteString("}\n")
	return buf.Bytes(), nil
}
