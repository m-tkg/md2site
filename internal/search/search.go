// Package search emits the client-side search index as a plain <script>
// file so search works over file:// (no fetch, no CORS).
package search

import (
	"encoding/json"
	"fmt"
)

type Entry struct {
	Title string `json:"t"`
	URL   string `json:"u"` // output-relative, e.g. "docs/setup/index.html"
	Body  string `json:"b"` // normalized plain text
}

// maxBodyRunes caps the indexed text per page to keep the index small.
const maxBodyRunes = 20000

// IndexJS renders the search index as JavaScript.
func IndexJS(entries []Entry) ([]byte, error) {
	for i := range entries {
		r := []rune(entries[i].Body)
		if len(r) > maxBodyRunes {
			entries[i].Body = string(r[:maxBodyRunes])
		}
	}
	data, err := json.Marshal(struct {
		Pages []Entry `json:"pages"`
	}{entries})
	if err != nil {
		return nil, fmt.Errorf("marshal search index: %w", err)
	}
	return append(append([]byte("window.__MD2SITE_INDEX__="), data...), byte('\n')), nil
}
