// Package scan walks an input directory and collects markdown files,
// applying fixed and user-supplied exclusion rules.
package scan

import (
	"io/fs"
	"path"
	"sort"
	"strings"
)

var fixedExcludedDirs = map[string]bool{
	"node_modules": true,
	"vendor":       true,
	".git":         true,
}

// Markdown returns the slash-separated paths (relative to root) of all
// markdown files under root, sorted. Hidden directories, node_modules,
// vendor and .git are always skipped; extra glob patterns are matched
// against both the relative path and the base name.
func Markdown(fsys fs.FS, excludes []string) ([]string, error) {
	var files []string
	err := fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if p == "." {
			return nil
		}
		name := d.Name()
		if d.IsDir() {
			if fixedExcludedDirs[name] || strings.HasPrefix(name, ".") || matchAny(excludes, p, name) {
				return fs.SkipDir
			}
			return nil
		}
		if strings.HasPrefix(name, ".") || !strings.EqualFold(path.Ext(name), ".md") {
			return nil
		}
		if matchAny(excludes, p, name) {
			return nil
		}
		files = append(files, p)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	return files, nil
}

func matchAny(patterns []string, relPath, base string) bool {
	for _, pat := range patterns {
		if ok, _ := path.Match(pat, relPath); ok {
			return true
		}
		if ok, _ := path.Match(pat, base); ok {
			return true
		}
		// Also allow patterns like "docs/drafts" to exclude a subtree.
		if relPath == pat || strings.HasPrefix(relPath, pat+"/") {
			return true
		}
	}
	return false
}
