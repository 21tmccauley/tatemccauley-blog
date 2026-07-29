package main

import (
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/ansi"
	"github.com/charmbracelet/glamour/styles"
	"github.com/charmbracelet/lipgloss"
)

// The SSH edition takes over the whole terminal pane and draws a window card on
// it: the pane gets a flat desktop fill, and the card sits centered on top with
// a dim border, a slightly lighter interior, and a titlebar with three window
// buttons. Type is a near-monochrome ramp, so the buttons and syntax
// highlighting are the only color in the UI. AdaptiveColor resolves
// per-renderer, so every value tracks the client terminal's light/dark
// background.
// The card and code fills are deliberately exact xterm-256 grays (#eeeeee is
// 255, #121212 is 233, and so on). Chroma quantizes a code block's background
// to 256 colors even when the session is truecolor, so an off-ramp value would
// leave a faint seam where highlighted code meets the card.
var (
	deskBg   = lipgloss.AdaptiveColor{Light: "#e4e4e4", Dark: "#080808"} // pane behind the card
	cardBg   = lipgloss.AdaptiveColor{Light: "#eeeeee", Dark: "#121212"}
	codeBg   = lipgloss.AdaptiveColor{Light: "#e4e4e4", Dark: "#262626"}
	lineCol  = lipgloss.AdaptiveColor{Light: "#d8d8e0", Dark: "#2b2b31"} // borders, rules, scroll track
	titleCol = lipgloss.AdaptiveColor{Light: "#15151b", Dark: "#eeeeee"} // page and section titles
	selCol   = lipgloss.AdaptiveColor{Light: "#000000", Dark: "#ffffff"} // selected row, active tab
	fgCol    = lipgloss.AdaptiveColor{Light: "#33333d", Dark: "#cbcbd2"} // body copy
	mutedCol = lipgloss.AdaptiveColor{Light: "#70707a", Dark: "#8a8a93"} // hints, tagline
	faintCol = lipgloss.AdaptiveColor{Light: "#a5a5af", Dark: "#5c5c65"} // window title, dates
	thumbCol = lipgloss.AdaptiveColor{Light: "#8a8a95", Dark: "#a0a0aa"} // scrollbar thumb
)

// The three window buttons are the one splash of color and stay fixed in both
// themes, the way the real ones do.
var (
	btnCloseCol = lipgloss.Color("#ff5f5f")
	btnMinCol   = lipgloss.Color("#ffaf00")
	btnMaxCol   = lipgloss.Color("#00d75f")
)

// theme holds every style bound to one renderer. Over SSH the renderer is the
// client's session (see serve.go), so styles must be built from it — package
// globals would resolve colors against the server's stdout and render plain.
type theme struct {
	r *lipgloss.Renderer

	card                     lipgloss.Style
	fill                     lipgloss.Style
	winTitle                 lipgloss.Style
	btnClose, btnMin, btnMax lipgloss.Style
	brand, tagline, section  lipgloss.Style
	body, muted, divider     lipgloss.Style
	hint                     lipgloss.Style
	itemTitle, itemDate      lipgloss.Style
	selTitle, selMarker      lipgloss.Style
	tabActive, tabInactive   lipgloss.Style
	thumb, track             lipgloss.Style
	barFilled, barEmpty      lipgloss.Style
	center                   lipgloss.Style
}

func newTheme(r *lipgloss.Renderer) *theme {
	t := &theme{r: r}

	// Every style used inside the card carries the card background. A style
	// that omitted it would punch a hole in the fill: lipgloss closes each
	// Render with a full reset, so any run that does not re-declare the
	// background falls back to the terminal's own.
	base := r.NewStyle().Background(cardBg)

	t.card = r.NewStyle().
		Background(cardBg).
		Border(lipgloss.NormalBorder()).
		BorderForeground(lineCol).
		BorderBackground(cardBg)
	t.fill = base
	t.winTitle = base.Foreground(faintCol)
	t.btnClose = base.Foreground(btnCloseCol)
	t.btnMin = base.Foreground(btnMinCol)
	t.btnMax = base.Foreground(btnMaxCol)
	t.brand = base.Bold(true).Foreground(titleCol)
	t.tagline = base.Foreground(mutedCol).Italic(true)
	t.section = base.Bold(true).Foreground(titleCol)
	t.body = base.Foreground(fgCol)
	t.muted = base.Foreground(mutedCol)
	t.divider = base.Foreground(lineCol)
	t.hint = base.Foreground(mutedCol)
	t.itemTitle = base.Foreground(fgCol)
	t.itemDate = base.Foreground(faintCol)
	t.selTitle = base.Bold(true).Foreground(selCol)
	t.selMarker = base.Bold(true).Foreground(selCol)
	t.tabActive = base.Bold(true).Foreground(selCol)
	t.tabInactive = base.Foreground(mutedCol)
	t.thumb = base.Foreground(thumbCol)
	t.track = base.Foreground(lineCol)
	t.barFilled = base.Foreground(fgCol)
	t.barEmpty = base.Foreground(lineCol)
	t.center = base.Align(lipgloss.Center)
	return t
}

// hex resolves an adaptive color for the renderer's background. glamour takes
// colors as plain strings, so it cannot use AdaptiveColor itself.
func (t *theme) hex(c lipgloss.AdaptiveColor) string {
	if t.r.HasDarkBackground() {
		return c.Dark
	}
	return c.Light
}

// pad returns n columns of card background, for spacing between styled runs.
// Raw spaces would show the terminal background through the card.
func (t *theme) pad(n int) string {
	if n <= 0 {
		return ""
	}
	return t.fill.Render(strings.Repeat(" ", n))
}

// rule renders a horizontal divider of the given width.
func (t *theme) rule(width int) string {
	if width < 1 {
		width = 1
	}
	return t.divider.Render(strings.Repeat("─", width))
}

// hints joins "key desc  •  key desc". Keys are not highlighted: the whole
// footer is one dim gray so it recedes. Styling and wrapping happen in
// model.footer, which knows the card width.
func (t *theme) hints(pairs ...[2]string) string {
	parts := make([]string, 0, len(pairs))
	for _, p := range pairs {
		parts = append(parts, p[0]+" "+p[1])
	}
	return strings.Join(parts, "  •  ")
}

// markdownRenderer builds a glamour renderer wrapped to width, matching the
// renderer's color profile and background so posts render in color over SSH.
func (t *theme) markdownRenderer(width int) (*glamour.TermRenderer, error) {
	if width < 20 {
		width = 20
	}
	name := "dark"
	if !t.r.HasDarkBackground() {
		name = "light"
	}

	opts := []glamour.TermRendererOption{}
	if base := styles.DefaultStyles[name]; base != nil {
		opts = append(opts, glamour.WithStyles(t.markdownStyle(*base)))
	} else {
		opts = append(opts, glamour.WithStandardStyle(name))
	}
	opts = append(opts,
		glamour.WithColorProfile(t.r.ColorProfile()),
		glamour.WithWordWrap(width),
	)
	return glamour.NewTermRenderer(opts...)
}

// markdownStyle adapts a base glamour style to the card: a flush-left document
// painted with the card background, and headings that are bold and underlined
// in body color rather than tinted, since the palette saves color for the
// window buttons.
//
// sc arrives as a shallow copy of a shared styles.DefaultStyles entry, so any
// pointer it holds still aims at library-global memory. Pointer fields are
// replaced rather than written through, and Chroma — the one nested pointer —
// is cloned below. Getting this wrong is a data race, not just untidiness:
// every SSH session builds its own renderer concurrently.
func (t *theme) markdownStyle(sc ansi.StyleConfig) ansi.StyleConfig {
	var (
		zero     = uint(0)
		yes      = true
		body     = t.hex(fgCol)
		muted    = t.hex(mutedCol)
		title    = t.hex(titleCol)
		line     = t.hex(lineCol)
		fill     = t.hex(cardBg)
		codeFill = t.hex(codeBg)
	)

	// Painting the document background is what keeps rendered posts on the same
	// fill as the chrome around them: glamour re-declares it after each of its
	// own resets, which an outer lipgloss style cannot do.
	sc.Document.Margin = &zero
	sc.Document.BackgroundColor = &fill
	sc.Document.Color = &body

	for _, h := range []*ansi.StyleBlock{&sc.H1, &sc.H2, &sc.H3, &sc.H4, &sc.H5, &sc.H6} {
		h.Prefix = ""
		h.BackgroundColor = &fill
		h.Color = &title
		h.Bold = &yes
		h.Underline = &yes
	}

	// The stock themes clear the background on a fenced block, which would cut a
	// hole in the card. A raised fill is no good either: glamour does not pad
	// code lines out to the wrap width, so a panel would end ragged on every
	// line. Keep fenced blocks on the card fill and let their indent and syntax
	// colors set them apart. Inline code still gets a chip — those spans are
	// short enough that a ragged edge is the point.
	sc.CodeBlock.BackgroundColor = &fill
	if sc.CodeBlock.Chroma != nil {
		chroma := *sc.CodeBlock.Chroma // clone before touching: shared global
		chroma.Text.BackgroundColor = &fill
		chroma.Background.BackgroundColor = &fill
		sc.CodeBlock.Chroma = &chroma
	}
	sc.Code.BackgroundColor = &codeFill
	sc.Code.Color = &body

	sc.Link.Color = &muted
	sc.Link.Underline = &yes
	sc.LinkText.Color = &body
	sc.LinkText.Bold = &yes
	sc.HorizontalRule.Color = &line
	sc.BlockQuote.Color = &muted
	sc.BlockQuote.Italic = &yes
	return sc
}
