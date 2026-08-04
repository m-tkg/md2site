package render

import (
	"strconv"
	"strings"
	"unicode"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
)

// unicodeIDs generates GitHub-style heading anchors that keep non-ASCII
// letters (goldmark's default generator reduces Japanese headings to the
// useless "heading", "heading-1", ...).
type unicodeIDs struct {
	used map[string]bool
}

func newIDs() parser.IDs {
	return &unicodeIDs{used: map[string]bool{}}
}

func (s *unicodeIDs) Generate(value []byte, kind ast.NodeKind) []byte {
	var b strings.Builder
	for _, r := range string(value) {
		switch {
		case unicode.IsLetter(r) || unicode.IsNumber(r):
			b.WriteRune(unicode.ToLower(r))
		case r == ' ' || r == '-' || r == '_':
			b.WriteByte('-')
		}
	}
	id := strings.Trim(b.String(), "-")
	if id == "" {
		id = "heading"
	}
	uniq := id
	for n := 1; s.used[uniq]; n++ {
		uniq = id + "-" + strconv.Itoa(n)
	}
	s.used[uniq] = true
	return []byte(uniq)
}

func (s *unicodeIDs) Put(value []byte) {
	s.used[string(value)] = true
}
