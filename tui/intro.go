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

func (m *model) viewIntro() string {
	frac := float64(m.frame) / float64(introFrames)
	if frac > 1 {
		frac = 1
	}

	box := m.th.introBox.Render(
		m.th.brand.Render(letterSpace("TATE McCAULEY")) + "\n" +
			m.th.tagline.Render("the terminal edition"))

	// Size the bar to just inside the box so it centers cleanly underneath.
	// The percentage lives on the status line, not after the bar — a trailing
	// label there would shift the bar's visual center off to the left.
	barWidth := lipgloss.Width(box) - 2
	if barWidth < 20 {
		barWidth = 20
	}
	status := m.th.key.Render(introSpinner[m.frame%len(introSpinner)]) + " " +
		m.th.muted.Render("establishing secure channel") +
		m.th.muted.Render(fmt.Sprintf("   %3.0f%%", frac*100))

	block := lipgloss.JoinVertical(lipgloss.Center,
		box,
		"",
		status,
		m.renderBar(barWidth, frac),
	)
	return m.th.r.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, block)
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
