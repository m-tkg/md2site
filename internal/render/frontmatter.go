package render

import (
	"bytes"

	"gopkg.in/yaml.v3"
)

type frontMatter struct {
	Title string `yaml:"title"`
}

// stripFrontMatter removes a leading YAML front matter block and returns its
// parsed fields plus the remaining markdown body.
func stripFrontMatter(src []byte) (frontMatter, []byte, bool) {
	lines := bytes.Split(src, []byte("\n"))
	if len(lines) == 0 || !bytes.Equal(bytes.TrimSpace(lines[0]), []byte("---")) {
		return frontMatter{}, src, false
	}
	end := -1
	for i := 1; i < len(lines); i++ {
		if bytes.Equal(bytes.TrimSpace(lines[i]), []byte("---")) {
			end = i
			break
		}
	}
	if end < 0 {
		return frontMatter{}, src, false
	}
	var fm frontMatter
	_ = yaml.Unmarshal(bytes.Join(lines[1:end], []byte("\n")), &fm)
	body := bytes.Join(lines[end+1:], []byte("\n"))
	return fm, body, true
}

func (d *Doc) mdSource() []byte {
	if len(d.body) > 0 {
		return d.body
	}
	return d.page.Source
}
