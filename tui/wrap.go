package main

import (
	"strings"

	"github.com/mattn/go-runewidth"
)

// wrapText breaks s into lines that fit within width (runewidth-aware).
// Newlines in s are preserved as paragraph breaks.
func wrapText(s string, width int) string {
	if width < 4 {
		width = 4
	}
	var b strings.Builder
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		line = strings.TrimRight(line, "\r")
		if line == "" {
			b.WriteByte('\n')
			continue
		}
		b.WriteString(wrapLine(line, width))
		if i < len(lines)-1 {
			b.WriteByte('\n')
		}
	}
	return strings.TrimSuffix(b.String(), "\n")
}

func wrapLine(line string, width int) string {
	if runewidth.StringWidth(line) <= width {
		return line
	}
	words := strings.Fields(line)
	if len(words) == 0 {
		return ""
	}
	var b strings.Builder
	lineW := 0
	writeWord := func(word string) {
		wl := runewidth.StringWidth(word)
		if wl > width {
			if lineW > 0 {
				b.WriteByte('\n')
				lineW = 0
			}
			for _, r := range word {
				rw := runewidth.RuneWidth(r)
				if rw <= 0 {
					rw = 1
				}
				if lineW+rw > width && lineW > 0 {
					b.WriteByte('\n')
					lineW = 0
				}
				b.WriteRune(r)
				lineW += rw
			}
			return
		}
		need := wl
		if lineW > 0 {
			need++
		}
		if lineW+need > width {
			b.WriteByte('\n')
			lineW = 0
		}
		if lineW > 0 {
			b.WriteByte(' ')
			lineW++
		}
		b.WriteString(word)
		lineW += wl
	}
	for _, w := range words {
		writeWord(w)
	}
	return b.String()
}
