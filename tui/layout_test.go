package main

import (
	"strings"
	"sync"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour/styles"
	"github.com/charmbracelet/lipgloss"
)

// TestCardFitsTerminal is the invariant the whole card layout rests on: the
// window must never render wider or taller than the terminal it was given. The
// card reserves rows for its own chrome and the footer wraps on a narrow
// column, so a miscount there pushes the border off screen — visible breakage
// that is easy to introduce and easy to miss by eye at one size.
func TestCardFitsTerminal(t *testing.T) {
	sizes := []struct{ w, h int }{
		{40, 16}, {56, 20}, {72, 24}, {80, 24}, {100, 30}, {120, 40}, {200, 50},
	}
	screens := []struct {
		name  string
		setup func(*model)
	}{
		{"intro", func(m *model) { m.screen = screenIntro }},
		{"home", func(m *model) { m.screen, m.tab = screenMain, tabHome }},
		{"blog", func(m *model) { m.screen, m.tab = screenMain, tabBlog }},
		{"page", func(m *model) { m.screen, m.tab = screenMain, tabResume }},
		{"post", func(m *model) { m.screen = screenPost }},
	}

	for _, sc := range screens {
		for _, size := range sizes {
			m := testModel()
			sc.setup(m)
			m.setSize(size.w, size.h)

			out := m.View()
			lines := strings.Split(out, "\n")
			if len(lines) > size.h {
				t.Errorf("%s at %dx%d: %d rows, want <= %d",
					sc.name, size.w, size.h, len(lines), size.h)
			}
			for i, line := range lines {
				if w := lipgloss.Width(line); w > size.w {
					t.Errorf("%s at %dx%d: row %d is %d wide, want <= %d",
						sc.name, size.w, size.h, i, w, size.w)
				}
			}
		}
	}
}

// TestScrollbarKeepsColumnWidth checks that every body row is the same width
// once the gutter is appended, whatever the thumb is doing. A row that came out
// short or long would ripple into the border.
func TestScrollbarKeepsColumnWidth(t *testing.T) {
	m := testModel()
	m.screen, m.tab = screenMain, tabResume
	m.setSize(80, 24)

	for _, offset := range []int{0, 1, 5, 1000} {
		m.vp.SetYOffset(offset)
		for i, line := range strings.Split(m.scrollbar(m.vp.View()), "\n") {
			if w := lipgloss.Width(line); w != m.contentWidth {
				t.Errorf("offset %d row %d: width %d, want %d",
					offset, i, w, m.contentWidth)
			}
		}
	}
}

// TestStripLeadingH1 covers the rule that keeps the post screen from showing its
// title twice: the card pins the front-matter title, so the body's opening
// heading has to go.
func TestStripLeadingH1(t *testing.T) {
	cases := []struct{ in, want string }{
		{"# A Title\n\nbody text", "body text"},
		{"body with no heading", "body with no heading"},
		{"## Not an H1\n\nbody", "## Not an H1\n\nbody"},
		{"# Only a heading", ""},
		{"#NoSpace\n\nbody", "#NoSpace\n\nbody"},
	}
	for _, c := range cases {
		if got := stripLeadingH1(c.in); got != c.want {
			t.Errorf("stripLeadingH1(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// TestBodyLinesFillTextColumn guards the card's background fill. Two things
// downstream pad short lines with *unstyled* space — viewport.View renders
// through a background-less lipgloss style, and JoinVertical pads with bare
// spaces — so any body line left short of the text column shows the visitor's
// terminal through the middle of the window. This was live on every post-list
// row before padBlock existed.
func TestBodyLinesFillTextColumn(t *testing.T) {
	screens := []struct {
		name  string
		setup func(*model)
	}{
		{"home", func(m *model) { m.screen, m.tab = screenMain, tabHome }},
		{"blog", func(m *model) { m.screen, m.tab = screenMain, tabBlog }},
		{"page", func(m *model) { m.screen, m.tab = screenMain, tabResume }},
		{"post", func(m *model) { m.screen = screenPost }},
	}
	for _, sc := range screens {
		for _, size := range []struct{ w, h int }{{40, 16}, {80, 24}, {120, 40}} {
			m := testModel()
			sc.setup(m)
			m.setSize(size.w, size.h)
			for i, line := range strings.Split(m.bodyFor(), "\n") {
				if w := lipgloss.Width(line); w != m.textWidth() {
					t.Errorf("%s at %dx%d: body line %d is %d wide, want %d",
						sc.name, size.w, size.h, i, w, m.textWidth())
				}
			}
		}
	}
}

// TestChromeLinesFillContentWidth is the same invariant for the rows the card
// composes outside the viewport. titleBar in particular drops the site name on a
// narrow card and used to return a 5-column line.
func TestChromeLinesFillContentWidth(t *testing.T) {
	for _, w := range []int{28, 30, 34, 36, 40, 80, 120} {
		m := testModel()
		m.screen = screenMain
		m.setSize(w, 30)

		blocks := map[string]string{
			"titleBar": m.titleBar(),
			"tabRow":   m.tabRow(),
			"footer":   m.footer(m.th.hints(blogHints...)),
		}
		for name, block := range blocks {
			for i, line := range strings.Split(block, "\n") {
				if got := lipgloss.Width(line); got != m.contentWidth {
					t.Errorf("w=%d %s line %d is %d wide, want %d",
						w, name, i, got, m.contentWidth)
				}
			}
		}
	}
}

// TestFooterFillsItsReservation keeps the card's bottom border from jumping a
// row as you move between tabs: bodyHeight spends the full hintLines
// reservation, so footer has to occupy all of it.
func TestFooterFillsItsReservation(t *testing.T) {
	for _, w := range []int{40, 56, 80, 120} {
		m := testModel()
		m.screen = screenMain
		m.setSize(w, 24)
		for _, set := range hintSets() {
			if got := lipgloss.Height(m.footer(m.th.hints(set...))); got != m.hintLines {
				t.Errorf("w=%d: footer rendered %d rows, reserved %d", w, got, m.hintLines)
			}
		}
	}
}

// TestTabScrollSurvivesShortIntermediateTab covers a scroll position being
// silently thrown away. viewport.SetYOffset clamps against the content it
// currently holds, so restoring an offset before loading the new content
// truncated it to zero whenever the tab being left happened to fit on one
// screen — which Home almost always does.
func TestTabScrollSurvivesShortIntermediateTab(t *testing.T) {
	m := testModel()
	m.pageMD[tabResume] = testLongPage(200)
	m.screen, m.tab = screenMain, tabResume
	m.setSize(92, 40)

	m.vp.SetYOffset(40)
	want := m.vp.YOffset
	if want == 0 {
		t.Fatal("setup: expected a scrollable page")
	}

	m.switchTab(tabHome) // Home fits entirely, so maxYOffset is 0 there
	m.switchTab(tabResume)
	if got := m.vp.YOffset; got != want {
		t.Errorf("offset after a round trip through a short tab = %d, want %d", got, want)
	}
}

// TestBlogKeysKeepCursorVisible covers the post list desyncing from its
// selection. Forwarding unhandled keys to the viewport let its own
// pgdown/space/d bindings scroll the list while the cursor stayed put, so enter
// opened a post that was nowhere on screen.
func TestBlogKeysKeepCursorVisible(t *testing.T) {
	keys := []string{"pgdown", " ", "end", "pgup", "home", "down", "up", "j", "k", "d", "u", "b", "f"}
	for _, size := range []struct{ w, h int }{{92, 18}, {92, 40}, {40, 16}} {
		for _, key := range keys {
			m := testModel()
			m.posts = make([]post, 30)
			for i := range m.posts {
				m.posts[i] = post{title: "a post title"}
			}
			m.screen, m.tab = screenMain, tabBlog
			m.setSize(size.w, size.h)

			m.updateMain(testKey(key))
			if m.cursor < m.vp.YOffset || m.cursor >= m.vp.YOffset+m.vp.Height {
				t.Errorf("%dx%d %q: cursor %d outside view [%d,%d)",
					size.w, size.h, key, m.cursor, m.vp.YOffset, m.vp.YOffset+m.vp.Height)
			}
		}
	}
}

// TestThemeDoesNotMutateGlamourGlobals covers a cross-session data race. A
// glamour ansi.StyleConfig copied out of styles.DefaultStyles is a shallow copy,
// and its Chroma field is a pointer into library-global memory. Every SSH
// session builds its own theme concurrently, so writing through that pointer is
// an unsynchronized write to shared state.
func TestThemeDoesNotMutateGlamourGlobals(t *testing.T) {
	for _, name := range []string{"dark", "light"} {
		base := styles.DefaultStyles[name]
		if base == nil || base.CodeBlock.Chroma == nil {
			continue
		}
		before := base.CodeBlock.Chroma.Text.BackgroundColor
		th := newTheme(lipgloss.DefaultRenderer())
		if _, err := th.markdownRenderer(60); err != nil {
			t.Fatal(err)
		}
		if after := base.CodeBlock.Chroma.Text.BackgroundColor; before != after {
			t.Errorf("%s: global Chroma mutated, before=%v after=%v", name, before, after)
		}
	}
}

// TestConcurrentThemeBuild is the above under -race, mirroring several SSH
// sessions connecting at once.
func TestConcurrentThemeBuild(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 24; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			th := newTheme(lipgloss.DefaultRenderer())
			if _, err := th.markdownRenderer(60); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
}

func testLongPage(n int) string {
	var b strings.Builder
	for i := 0; i < n; i++ {
		b.WriteString("a line of prose long enough to wrap somewhere\n\n")
	}
	return b.String()
}

func testKey(s string) tea.KeyMsg {
	switch s {
	case " ":
		return tea.KeyMsg{Type: tea.KeySpace}
	case "pgdown":
		return tea.KeyMsg{Type: tea.KeyPgDown}
	case "pgup":
		return tea.KeyMsg{Type: tea.KeyPgUp}
	case "end":
		return tea.KeyMsg{Type: tea.KeyEnd}
	case "home":
		return tea.KeyMsg{Type: tea.KeyHome}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
	}
}
