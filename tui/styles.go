package main

import (
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/ansi"
	"github.com/charmbracelet/glamour/styles"
	"github.com/charmbracelet/lipgloss"
)

// Brand palette mirrored from tatemccauley.com so the SSH edition reads as the
// same site. AdaptiveColor resolves per-renderer, so these track the client
// terminal's light/dark background.
var (
	accentColor = lipgloss.AdaptiveColor{Light: "#0011ff", Dark: "#5588ff"}
	fgColor     = lipgloss.AdaptiveColor{Light: "#232333", Dark: "#d0d0d0"}
	mutedColor  = lipgloss.AdaptiveColor{Light: "#555555", Dark: "#999999"}
	subtleColor = lipgloss.AdaptiveColor{Light: "#c2c2cc", Dark: "#3a3a3a"}
)

// theme holds every style bound to one renderer. Over SSH the renderer is the
// client's session (see serve.go), so styles must be built from it — package
// globals would resolve colors against the server's stdout and render plain.
type theme struct {
	r *lipgloss.Renderer

	brand, tagline, section lipgloss.Style
	body, muted, divider    lipgloss.Style
	splashBox, center       lipgloss.Style
	key, desc               lipgloss.Style
	itemTitle, itemDate     lipgloss.Style
	selTitle, selMarker     lipgloss.Style
	tabActive, tabInactive  lipgloss.Style
	introBox                lipgloss.Style
	barFilled, barEmpty     lipgloss.Style
}

func newTheme(r *lipgloss.Renderer) *theme {
	t := &theme{r: r}
	t.brand = r.NewStyle().Bold(true).Foreground(accentColor)
	t.tagline = r.NewStyle().Foreground(mutedColor).Italic(true)
	t.section = r.NewStyle().Bold(true).Foreground(mutedColor)
	t.body = r.NewStyle().Foreground(fgColor)
	t.muted = r.NewStyle().Foreground(mutedColor)
	t.divider = r.NewStyle().Foreground(subtleColor)
	t.splashBox = r.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(accentColor).
		Padding(0, 3).
		Align(lipgloss.Center)
	t.center = r.NewStyle().Align(lipgloss.Center)
	t.key = r.NewStyle().Bold(true).Foreground(accentColor)
	t.desc = r.NewStyle().Foreground(mutedColor)
	t.itemTitle = r.NewStyle().Foreground(fgColor)
	t.itemDate = r.NewStyle().Foreground(mutedColor)
	t.selTitle = r.NewStyle().Bold(true).Foreground(accentColor)
	t.selMarker = r.NewStyle().Bold(true).Foreground(accentColor)
	t.tabActive = r.NewStyle().Bold(true).Foreground(accentColor).Underline(true)
	t.tabInactive = r.NewStyle().Foreground(mutedColor)
	t.introBox = r.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(accentColor).
		Padding(1, 4).
		Align(lipgloss.Center)
	t.barFilled = r.NewStyle().Foreground(accentColor)
	t.barEmpty = r.NewStyle().Foreground(subtleColor)
	return t
}

// rule renders a horizontal divider of the given width.
func (t *theme) rule(width int) string {
	if width < 1 {
		width = 1
	}
	return t.divider.Render(strings.Repeat("─", width))
}

// hints renders "key desc   ·   key desc", accent keys and muted descriptions.
func (t *theme) hints(pairs ...[2]string) string {
	parts := make([]string, 0, len(pairs))
	for _, p := range pairs {
		parts = append(parts, t.key.Render(p[0])+" "+t.desc.Render(p[1]))
	}
	return strings.Join(parts, t.desc.Render("   ·   "))
}

// markdownRenderer builds a glamour renderer wrapped to width, matching the
// renderer's color profile and background so posts render in color over SSH.
func (t *theme) markdownRenderer(width int) (*glamour.TermRenderer, error) {
	if width < 20 {
		width = 20
	}
	name, accent := "dark", "#5588ff"
	if !t.r.HasDarkBackground() {
		name, accent = "light", "#0011ff"
	}

	opts := []glamour.TermRendererOption{}
	if base := styles.DefaultStyles[name]; base != nil {
		opts = append(opts, glamour.WithStyles(brandMarkdownStyle(*base, accent)))
	} else {
		opts = append(opts, glamour.WithStandardStyle(name))
	}
	opts = append(opts,
		glamour.WithColorProfile(t.r.ColorProfile()),
		glamour.WithWordWrap(width),
	)
	return glamour.NewTermRenderer(opts...)
}

// brandMarkdownStyle tweaks a base glamour style for the blog: a flush-left
// document (so body aligns with the chrome) and blue, prefix-free headings and
// links matching the site accent. Pointer fields are replaced, never mutated
// in place, so the shared DefaultStyles entry is untouched.
func brandMarkdownStyle(sc ansi.StyleConfig, accent string) ansi.StyleConfig {
	zero := uint(0)
	bold := true
	sc.Document.Margin = &zero
	for _, h := range []*ansi.StyleBlock{&sc.H1, &sc.H2, &sc.H3, &sc.H4, &sc.H5, &sc.H6} {
		h.Prefix = ""
		h.BackgroundColor = nil
		h.Color = &accent
		h.Bold = &bold
	}
	sc.Link.Color = &accent
	return sc
}
