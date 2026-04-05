package main

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// loadPosts reads every *.md file in dir except index.md (Eleventy listing page).
// title is the filename without .md; body is the full raw file contents.
func loadPosts(dir string) ([]post, error) {
	paths, err := filepath.Glob(filepath.Join(dir, "*.md"))
	if err != nil {
		return nil, err
	}
	slices.Sort(paths)

	var posts []post
	for _, path := range paths {
		base := filepath.Base(path)
		if base == "index.md" {
			continue
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		title := strings.TrimSuffix(base, filepath.Ext(base))
		posts = append(posts, post{
			title: title,
			body:  string(b),
		})
	}
	return posts, nil
}

// stripFrontMatter removes a leading YAML block delimited by --- lines.
func stripFrontMatter(content string) string {
	s := strings.TrimSpace(content)
	if !strings.HasPrefix(s, "---") {
		return s
	}
	rest := strings.TrimPrefix(s, "---")
	rest = strings.TrimPrefix(rest, "\n")
	sep := "\n---"
	i := strings.Index(rest, sep)
	if i < 0 {
		return s
	}
	body := rest[i+len(sep):]
	return strings.TrimSpace(strings.TrimPrefix(body, "\n"))
}

// loadTUIHome reads the TUI-only home markdown (plain content, no Eleventy).
// Optional YAML front matter is stripped if present.
func loadTUIHome(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return stripFrontMatter(string(b)), nil
}
