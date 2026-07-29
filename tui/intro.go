package main

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// The intro is a brief branded "boot" shown on connect before the home screen —
// the first thing a visitor sees after `ssh tatemccauley.com`. It auto-advances
// after introFrames ticks; any key skips it.
const (
	introTickInterval = 90 * time.Millisecond
	introFrames       = 18 // ~1.7s, including one frame held at 100%
)

var introSpinner = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

type introTickMsg struct{}

func introTickCmd() tea.Cmd {
	return tea.Tick(introTickInterval, func(time.Time) tea.Msg { return introTickMsg{} })
}

// viewIntro paints the boot inside the same window card as the rest of the UI,
// so the card is on screen from the first frame and the home screen is a change
// of contents rather than a change of furniture.
func (m *model) viewIntro() string {
	frac := float64(m.frame) / float64(introFrames)
	if frac > 1 {
		frac = 1
	}

	cw := m.contentWidth
	wordmark := m.th.center.Width(cw).Render(m.th.brand.Render(letterSpace("TATE McCAULEY")))
	tagline := m.th.center.Width(cw).Render(m.th.tagline.Render("the terminal edition"))

	// The percentage sits on the status line rather than after the bar: a
	// trailing label there would shift the bar's visual center off to the left.
	barWidth := cw / 2
	if barWidth < 20 {
		barWidth = 20
	}
	if barWidth > cw {
		barWidth = cw
	}
	status := m.th.muted.Render(introSpinner[m.frame%len(introSpinner)]) + m.th.pad(1) +
		m.th.muted.Render("establishing secure channel") +
		m.th.muted.Render(fmt.Sprintf("   %3.0f%%", frac*100))

	blank := m.th.pad(cw)
	block := lipgloss.JoinVertical(lipgloss.Left,
		blank,
		wordmark,
		tagline,
		blank,
		m.th.center.Width(cw).Render(status),
		m.th.center.Width(cw).Render(m.renderBar(barWidth, frac)),
	)

	return m.onDesktop(m.th.card.
		Padding(cardPadY, m.padX).
		Width(cw + 2*m.padX).
		Render(lipgloss.JoinVertical(lipgloss.Left, m.titleBar(), block)))
}

func (m *model) renderBar(width int, frac float64) string {
	if frac < 0 {
		frac = 0
	}
	if frac > 1 {
		frac = 1
	}
	fill := int(frac * float64(width))
	if fill > width {
		fill = width
	}
	return m.th.barFilled.Render(strings.Repeat("█", fill)) +
		m.th.barEmpty.Render(strings.Repeat("░", width-fill))
}

// letterSpace widens a wordmark: single spaces between letters, wider gaps
// between words — "TATE McCAULEY" -> "T A T E   M c C A U L E Y".
func letterSpace(s string) string {
	var b strings.Builder
	for _, word := range strings.Fields(s) {
		if b.Len() > 0 {
			b.WriteString("   ")
		}
		for i, r := range []rune(word) {
			if i > 0 {
				b.WriteByte(' ')
			}
			b.WriteRune(r)
		}
	}
	return b.String()
}
