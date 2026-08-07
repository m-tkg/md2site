package render

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"html"
	"html/template"
	"io"
	"strings"
)

// parseTabular reads CSV or TSV bytes into rows. Empty lines are skipped.
func parseTabular(data []byte, tsv bool) ([][]string, error) {
	r := csv.NewReader(bytes.NewReader(data))
	r.FieldsPerRecord = -1
	r.LazyQuotes = true
	if tsv {
		r.Comma = '\t'
	}
	var rows [][]string
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if len(record) == 1 && strings.TrimSpace(record[0]) == "" {
			continue
		}
		rows = append(rows, record)
	}
	return rows, nil
}

// TableHTML renders tabular rows as an HTML table with a page title.
func TableHTML(title string, rows [][]string) (template.HTML, error) {
	var b strings.Builder
	fmt.Fprintf(&b, "<h1>%s</h1>\n", html.EscapeString(title))
	if len(rows) == 0 {
		b.WriteString(`<p class="tabular-empty">No rows.</p>`)
		return template.HTML(b.String()), nil
	}
	b.WriteString(`<div class="tabular-wrap"><table class="tabular-table">`)
	writeTableRow(&b, rows[0], true)
	if len(rows) > 1 {
		b.WriteString("<tbody>")
		for _, row := range rows[1:] {
			writeTableRow(&b, row, false)
		}
		b.WriteString("</tbody>")
	}
	b.WriteString("</table></div>")
	return template.HTML(b.String()), nil
}

func writeTableRow(b *strings.Builder, row []string, header bool) {
	if header {
		b.WriteString("<thead><tr>")
	} else {
		b.WriteString("<tr>")
	}
	tag := "td"
	if header {
		tag = "th"
	}
	for _, cell := range row {
		fmt.Fprintf(b, "<%s>%s</%s>", tag, html.EscapeString(cell), tag)
	}
	if header {
		b.WriteString("</tr></thead>")
	} else {
		b.WriteString("</tr>")
	}
}

func tablePlainText(rows [][]string) string {
	var parts []string
	for _, row := range rows {
		for _, cell := range row {
			if s := strings.TrimSpace(cell); s != "" {
				parts = append(parts, s)
			}
		}
	}
	return strings.Join(parts, " ")
}
