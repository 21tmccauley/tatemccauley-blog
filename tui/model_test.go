package main

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func testModel() *model {
	pages := map[tab]string{
		tabAbout:  "# About\n\nabout body",
		tabResume: "# Resume\n\nresume body",
		tabNow:    "# Now\n\nnow body",
	}
	return newModel(lipgloss.DefaultRenderer(), []post{{title: "P", body: "b"}}, "home body", pages)
}

// TestSetSizeClampsUntrustedDimensions guards against a forged window-change:
// SSH sessions are anonymous and the client controls the reported size, so a
// huge value must not reach the width/height that drive padding allocations.
func TestSetSizeClampsUntrustedDimensions(t *testing.T) {
	m := testModel()
	m.setSize(1<<30, 1<<30)
	if m.width != maxWidth || m.height != maxHeight {
		t.Errorf("dimensions = %dx%d, want clamped to %dx%d", m.width, m.height, maxWidth, maxHeight)
	}
	if m.contentWidth > maxColumn {
		t.Errorf("contentWidth = %d, want <= %d", m.contentWidth, maxColumn)
	}

	// A render at the clamped size must not blow up or emit absurd line widths.
	m.screen = screenMain
	out := m.View()
	for _, line := range strings.Split(out, "\n") {
		if w := lipgloss.Width(line); w > maxWidth {
			t.Fatalf("rendered line width %d exceeds clamp %d", w, maxWidth)
		}
	}

	// Negative dimensions (seen on some early resize events) clamp to zero.
	m.setSize(-5, -5)
	if m.width != 0 || m.height != 0 {
		t.Errorf("negative dims = %dx%d, want 0x0", m.width, m.height)
	}
}

// TestSetSizeSkipsRerenderWhenWidthUnchanged verifies a height-only resize
// reuses the cached glamour renders instead of recomputing them.
func TestSetSizeSkipsRerenderWhenWidthUnchanged(t *testing.T) {
	m := testModel()
	m.setSize(80, 24)
	renderer, homeView := m.md, m.homeView
	if renderer == nil {
		t.Fatal("expected a markdown renderer after first sizing")
	}

	m.setSize(80, 40) // same width, taller
	if m.md != renderer {
		t.Error("renderer was rebuilt on a height-only resize")
	}
	if m.homeView != homeView {
		t.Error("home view was re-rendered on a height-only resize")
	}
	if m.height != 40 {
		t.Errorf("height = %d, want 40", m.height)
	}

	m.setSize(120, 40) // width change forces a rebuild
	if m.md == renderer {
		t.Error("renderer was not rebuilt on a width change")
	}
}
