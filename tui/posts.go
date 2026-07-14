package main

import (
	"io/fs"
	"path"
	"slices"
	"strings"
	"time"
)

// loadPosts reads every posts/*.md file in fsys except index.md (the Eleventy
// listing page). Title and date come from the front matter, falling back to
// the filename when absent. Posts are returned newest first.
func loadPosts(fsys fs.FS) ([]post, error) {
	paths, err := fs.Glob(fsys, "posts/*.md")
	if err != nil {
		return nil, err
	}

	var posts []post
	for _, p := range paths {
		base := path.Base(p)
		if base == "index.md" {
			continue
		}
		b, err := fs.ReadFile(fsys, p)
		if err != nil {
			return nil, err
		}
		meta, body := splitFrontMatter(string(b))
		title, date := parseMeta(meta)
		if title == "" {
			title = strings.TrimSuffix(base, path.Ext(base))
		}
		posts = append(posts, post{title: title, date: date, body: body})
	}
	slices.SortStableFunc(posts, func(a, b post) int {
		return b.date.Compare(a.date)
	})
	return posts, nil
}

// splitFrontMatter separates a leading YAML block delimited by --- lines from
// the rest of the document. meta is "" when there is no block.
func splitFrontMatter(content string) (meta, body string) {
	s := strings.TrimSpace(content)
	if !strings.HasPrefix(s, "---") {
		return "", s
	}
	rest := strings.TrimPrefix(s, "---")
	rest = strings.TrimPrefix(rest, "\n")
	const sep = "\n---"
	i := strings.Index(rest, sep)
	if i < 0 {
		return "", s
	}
	body = rest[i+len(sep):]
	return rest[:i], strings.TrimSpace(strings.TrimPrefix(body, "\n"))
}

// parseMeta pulls title and date out of front matter lines. The blog's front
// matter is flat key: value pairs, so a YAML library would be overkill.
func parseMeta(meta string) (title string, date time.Time) {
	for _, line := range strings.Split(meta, "\n") {
		key, val, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		val = strings.Trim(strings.TrimSpace(val), `"'`)
		switch strings.TrimSpace(key) {
		case "title":
			title = val
		case "date":
			if t, err := time.Parse("2006-01-02", val); err == nil {
				date = t
			}
		}
	}
	return title, date
}

// loadHome reads the TUI-only home markdown, stripping optional front matter.
func loadHome(fsys fs.FS) (string, error) {
	b, err := fs.ReadFile(fsys, "tui/home.md")
	if err != nil {
		return "", err
	}
	_, body := splitFrontMatter(string(b))
	return body, nil
}
