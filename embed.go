// Package blog embeds the site content consumed by the SSH/terminal edition
// (see tui/). Living at the repo root lets go:embed reach posts/ and the
// top-level pages, which the Eleventy site shares; the tui binary is fully
// self-contained as a result.
package blog

import "embed"

//go:embed posts/*.md tui/home.md about.md now.md resume.md
var Content embed.FS
