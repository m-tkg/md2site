// Package scan walks an input directory and collects site source files,
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

var sourceExts = map[string]bool{
	".md":  true,
	".csv": true,
	".tsv": true,
}

// Sources returns the slash-separated paths (relative to root) of all
// supported source files under root, sorted. Hidden directories,
// node_modules, vendor and .git are always skipped; extra glob patterns are
// matched against both the relative path and the base name.
func Sources(fsys fs.FS, excludes []string) ([]string, error) {
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
		if strings.HasPrefix(name, ".") || !sourceExts[strings.ToLower(path.Ext(name))] {
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

// Markdown is an alias for Sources kept for compatibility.
func Markdown(fsys fs.FS, excludes []string) ([]string, error) {
	return Sources(fsys, excludes)
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
