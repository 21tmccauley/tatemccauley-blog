package main

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Card geometry. The card is centered in the terminal on both axes, capped so
// prose keeps a readable measure however wide the window is, and thinned before
// the text column shrinks when the terminal gets narrow.
const (
	maxColumn = 76 // widest content column inside the card
	// The floor covers the gutter as well as the text, so textWidth never drops
	// below the width glamour is willing to wrap at. If it did, glamour would
	// emit lines wider than the column and the viewport would chop them
	// mid-word, with no ellipsis to show it had happened.
	minColumn     = minTextColumn + gutterW
	minTextColumn = 20
	cardPadX      = 3 // interior padding, left and right
	cardPadY      = 1 // blank row inside the top and bottom border
	cardMargX     = 4 // desktop showing between card and terminal edge
	cardMargY     = 2
	gutterW       = 2 // scrollbar gutter: one space, one bar cell

	navLines       = 1 // the nav degrades to fit rather than wrapping
	postTitleLines = 2 // the post screen's pinned title, and the blank under it
	minBody        = 3 // rows the viewport keeps however short the terminal is
)

const siteName = "tatemccauley.com"

// cardLayout returns the content width and horizontal padding for a terminal
// of width w.
func cardLayout(w int) (contentWidth, padX int) {
	avail := w - 2*cardMargX
	if widest := maxColumn + 2*cardPadX + 2; avail > widest {
		avail = widest
	}
	// Out of room: give the text its columns back by thinning the padding before
	// narrowing the measure any further. Deciding this from the resulting width
	// rather than from w keeps the layout monotonic — picking the thin padding by
	// a width threshold instead made one extra terminal column cost three
	// columns of measure.
	padX = cardPadX
	contentWidth = avail - 2 - 2*padX
	if contentWidth < minColumn {
		padX = 1
		contentWidth = avail - 2 - 2*padX
	}
	if contentWidth < minColumn {
		contentWidth = minColumn
	}
	return contentWidth, padX
}

// textWidth is the width available to content once the scrollbar gutter is
// reserved. The gutter is reserved on every screen so text does not reflow when
// a view happens to fit without scrolling.
func (m *model) textWidth() int {
	w := m.contentWidth - gutterW
	if w < minTextColumn {
		w = minTextColumn
	}
	return w
}

// titleBar is the window chrome every screen shares: three buttons on the
// left, the site name on the right, and a rule beneath. On a card too narrow
// for both, the name is dropped so the buttons stay put.
func (m *model) titleBar() string {
	buttons := m.th.btnClose.Render("●") + m.th.pad(1) +
		m.th.btnMin.Render("●") + m.th.pad(1) +
		m.th.btnMax.Render("●")
	line := buttons
	title := m.th.winTitle.Render(siteName)
	if gap := m.contentWidth - lipgloss.Width(buttons) - lipgloss.Width(title); gap >= 2 {
		line = buttons + m.th.pad(gap) + title
	}
	// fillLine matters on the narrow path, where the line is only the buttons:
	// JoinVertical would pad the rest with unstyled space.
	return m.fillLine(line) + "\n" + m.th.rule(m.contentWidth)
}

// tabRow is the site's global nav — Home · Blog · About · Resume · Now — with
// the active entry bold. On the post screen Blog stays active, so the nav still
// says where you are. When the card is too narrow for the full row the spacing
// tightens, and then the row gives way to the current tab alone: either beats
// letting lipgloss wrap and break a name in half.
func (m *model) tabRow() string {
	row := func(gap int) string {
		items := make([]string, 0, tabCount)
		for i, title := range tabTitles {
			style := m.th.tabInactive
			if tab(i) == m.tab {
				style = m.th.tabActive
			}
			items = append(items, style.Render(title))
		}
		return strings.Join(items, m.th.pad(gap))
	}

	line := row(2)
	if lipgloss.Width(line) > m.contentWidth {
		line = row(1)
	}
	if lipgloss.Width(line) > m.contentWidth {
		line = m.th.tabActive.Render(tabTitles[m.tab])
	}
	return m.th.fill.Width(m.contentWidth).MaxHeight(navLines).Render(line)
}

// scrollbar renders content with a thumb in a right-hand gutter, sized and
// placed to match the visible slice. It pads every line to the full content
// width itself: lipgloss would pad with unstyled space, which shows the
// terminal background through the card.
func (m *model) scrollbar(content string) string {
	lines := strings.Split(content, "\n")
	total, visible := m.vp.TotalLineCount(), m.vp.VisibleLineCount()

	thumbStart, thumbEnd := -1, -1
	if h := len(lines); total > visible && h > 0 {
		size := int(math.Round(float64(visible) / float64(total) * float64(h)))
		if size < 1 {
			size = 1
		}
		if size > h {
			size = h
		}
		maxOff := total - visible
		off := m.vp.YOffset
		if off > maxOff {
			off = maxOff
		}
		if off < 0 {
			off = 0
		}
		start := 0
		if maxOff > 0 {
			start = int(math.Round(float64(off) / float64(maxOff) * float64(h-size)))
		}
		thumbStart, thumbEnd = start, start+size
	}

	textW := m.textWidth()
	out := make([]string, len(lines))
	for i, ln := range lines {
		var cell string
		switch {
		case thumbStart < 0:
			cell = m.th.pad(1) // nothing to scroll; hold the gutter open
		case i >= thumbStart && i < thumbEnd:
			cell = m.th.thumb.Render("┃")
		default:
			cell = m.th.track.Render("│")
		}
		// Belt and braces: viewport.View already truncates to vp.Width, so this
		// cannot fire for viewport content. It matters only if scrollbar is ever
		// handed a block from somewhere else — a line wider than the column would
		// push the gutter past the border and wrap the whole card.
		if lipgloss.Width(ln) > textW {
			ln = m.th.fill.MaxWidth(textW).Render(ln)
		}
		out[i] = ln + m.th.pad(textW-lipgloss.Width(ln)+1) + cell
	}
	return strings.Join(out, "\n")
}

// lineLabel reports the visible slice of a scrolling view, as
// "lines 30–55 of 55". It is empty when everything already fits.
func (m *model) lineLabel() string {
	total, visible := m.vp.TotalLineCount(), m.vp.VisibleLineCount()
	if total == 0 || total <= visible {
		return ""
	}
	last := m.vp.YOffset + visible
	if last > total {
		last = total
	}
	return fmt.Sprintf("lines %d–%d of %d", m.vp.YOffset+1, last, total)
}

// footer renders the key hints, and the scroll position when there is one.
//
// The block always occupies exactly the rows the layout reserved: Height pads
// it out so the card's bottom border does not jump by a row as you move between
// tabs with differently-sized footers, and MaxHeight is the backstop that keeps
// a long line from pushing the card past the bottom of the terminal. When the
// position label will not fit in those rows it is dropped rather than clipped —
// half a label reads as breakage, no label reads as a narrow window.
func (m *model) footer(hints string) string {
	style := m.th.hint.Width(m.contentWidth).Height(m.hintLines).MaxHeight(m.hintLines)
	if label := m.lineLabel(); label != "" {
		withLabel := hints + "  •  " + label
		if lipgloss.Height(m.th.hint.Width(m.contentWidth).Render(withLabel)) <= m.hintLines {
			hints = withLabel
		}
	}
	return style.Render(hints)
}

// fillLine pads a composed line out to the full content width, so lipgloss
// never has to pad it with unstyled space and let the terminal background show
// through the card.
func (m *model) fillLine(s string) string {
	return s + m.th.pad(m.contentWidth-lipgloss.Width(s))
}

// card frames content and hints in the window: titlebar, nav, body, footer.
func (m *model) card(content, hints string) string {
	blank := m.th.pad(m.contentWidth)
	inner := lipgloss.JoinVertical(lipgloss.Left,
		m.titleBar(),
		blank,
		m.tabRow(),
		blank,
		content,
		blank,
		m.footer(hints),
	)
	return m.onDesktop(m.th.card.
		Padding(cardPadY, m.padX).
		Width(m.contentWidth + 2*m.padX).
		Render(inner))
}

// onDesktop centers the card in the terminal and paints everything around it,
// so the app owns the whole pane rather than letting the visitor's wallpaper
// show through the margins. The card hugs its content, so this fill is most of
// the screen on a short view and a thin frame on a long one.
func (m *model) onDesktop(box string) string {
	return m.th.r.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box,
		lipgloss.WithWhitespaceBackground(deskBg))
}

// chromeBase is every row the card spends on itself on an ordinary screen.
func (m *model) chromeBase() int {
	const borders, titleBar, blanks = 2, 2, 3
	return borders + 2*cardPadY + titleBar + blanks + navLines + m.hintLines
}

// chromeHeight is chromeBase plus anything the current screen adds, so the
// viewport can claim the rest.
func (m *model) chromeHeight() int {
	if m.screen == screenPost {
		return m.chromeBase() + postTitleLines
	}
	return m.chromeBase()
}

// hintSets lists every footer the UI can show. The card reserves height for the
// tallest, so the layout does not shift as you move between tabs.
func hintSets() [][][2]string {
	return [][][2]string{
		homeHints, blogHints, pageHints, postHints,
	}
}

var (
	homeHints = [][2]string{{"←/→", "pages"}, {"enter", "posts"}, {"q", "quit"}}
	blogHints = [][2]string{{"↑/↓", "move"}, {"enter", "open"}, {"←/→", "pages"}, {"q", "quit"}}
	pageHints = [][2]string{{"↑/↓", "scroll"}, {"←/→", "pages"}, {"q", "quit"}}
	postHints = [][2]string{{"↑/↓", "scroll"}, {"esc", "back"}, {"←/→", "pages"}, {"q", "quit"}}
)

// measureChrome records how many rows the footer needs at the current width.
// It wraps on a narrow card and the viewport height depends on it, so it is
// measured once per resize rather than once per frame.
func (m *model) measureChrome() {
	// Measured with a worst-case line label, so a long post cannot wrap the
	// footer onto a row the layout never reserved.
	const widestLabel = "  •  lines 8888–8888 of 8888"
	m.hintLines = 1
	for _, set := range hintSets() {
		s := m.th.hints(set...) + widestLabel
		if h := lipgloss.Height(m.th.hint.Width(m.contentWidth).Render(s)); h > m.hintLines {
			m.hintLines = h
		}
	}

	// On a short terminal the footer yields rows before the body does: it is
	// clamped to hintLines when rendered, so trimming the reservation truncates
	// the hints instead of pushing the top of the card off the screen. The post
	// screen's extra rows are counted here too, so no screen overflows.
	for m.hintLines > 1 && m.height-2*cardMargY-m.chromeBase()-postTitleLines < minBody {
		m.hintLines--
	}
}
