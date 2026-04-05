package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type screen int

const (
	screenHome screen = iota
	screenList
	screenPost
)

// Lines reserved below the post viewport (key hint). Title + body live inside the viewport.
const postViewFooterLines = 2

type post struct {
	title string
	body string
}

type model struct {
	vp viewport.Model
	width, height int
	screen        screen
	posts         []post
	cursor        int
	homeCursor    int
	homeMarkdown  string // body from tui/home.md (optional front matter stripped)
}

func (m *model) Init() tea.Cmd {
	return nil
}

func (m *model) resizePostViewport() {
	if m.width <= 0 || m.height <= 0 {
		return
	}
	vh := m.height - postViewFooterLines
	if vh < 1 {
		vh = 1
	}
	vw := m.width
	if vw < 1 {
		vw = 1
	}
	m.vp.Width = vw
	m.vp.Height = vh
}

// postViewportMarkdown returns title + body wrapped to the current viewport width
// so lines are not truncated horizontally (viewport only splits on \n).
func (m *model) postViewportMarkdown() string {
	if m.cursor < 0 || m.cursor >= len(m.posts) {
		return ""
	}
	w := m.vp.Width
	if w < 4 {
		w = m.width
	}
	if w < 4 {
		w = 72
	}
	p := m.posts[m.cursor]
	return wrapText("Post: "+p.title, w) + "\n\n" + wrapText(p.body, w)
}

func (m *model) loadPostIntoViewport() {
	if m.cursor < 0 || m.cursor >= len(m.posts) {
		return
	}
	m.resizePostViewport()
	m.vp.SetContent(m.postViewportMarkdown())
	m.vp.GotoTop()
}

func (m *model) reloadPostViewportAfterResize() {
	if m.screen != screenPost || m.cursor < 0 || m.cursor >= len(m.posts) {
		return
	}
	prev := m.vp.YOffset
	m.vp.SetContent(m.postViewportMarkdown())
	m.vp.SetYOffset(prev)
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.resizePostViewport()
		m.reloadPostViewportAfterResize()
		return m, nil
	case tea.KeyMsg:
		switch m.screen {
		case screenHome:
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "enter", "p":
				m.screen = screenList
			}
			// Some terminals send Enter as KeyEnter (CR) without matching "enter" in edge cases.
			if msg.Type == tea.KeyEnter {
				m.screen = screenList
			}
			return m, nil
		case screenList:
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(m.posts)-1 {
					m.cursor++
				}
			case "enter":
				if len(m.posts) > 0 {
					m.screen = screenPost
					m.loadPostIntoViewport()
				}
			case "esc":
				m.screen = screenHome
			}
			if m.screen == screenList && msg.Type == tea.KeyEnter && len(m.posts) > 0 {
				m.screen = screenPost
				m.loadPostIntoViewport()
			}
		case screenPost:
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "esc":
				m.screen = screenList
				return m, nil
			}
			var cmd tea.Cmd
			m.vp, cmd = m.vp.Update(msg)
			return m, cmd
		}
		return m, nil
	}
	return m, nil
}

func (m *model) View() string {
	switch m.screen {
	case screenList:
		return m.viewList()
	case screenPost:
		return m.viewPost()
	case screenHome:
		return m.viewHome()
	default:
		return "Unknown screen"
	}
}

func (m *model) viewHome() string {
	var b strings.Builder
	if strings.TrimSpace(m.homeMarkdown) != "" {
		if m.width > 0 {
			b.WriteString(wrapText(m.homeMarkdown, m.width))
		} else {
			b.WriteString(m.homeMarkdown)
		}
	} else {
		b.WriteString("(home.md missing or empty)\n")
	}
	b.WriteString("\n\n—\n")
	if n := len(m.posts); n > 0 {
		b.WriteString("Recent posts:\n")
		max := 3
		if n < max {
			max = n
		}
		for i := 0; i < max; i++ {
			title := m.posts[i].title
			if m.width > 0 {
				title = wrapText("  • "+title, m.width)
			} else {
				title = "  • " + title
			}
			b.WriteString(title)
			b.WriteString("\n")
		}
	}
	hint := "\np / Enter — all posts • q — quit\n"
	if m.width > 0 {
		b.WriteString("\n")
		b.WriteString(wrapText(strings.TrimSpace(hint), m.width))
		b.WriteString("\n")
	} else {
		b.WriteString(hint)
	}
	return b.String()
}

func (m *model) viewList() string {
	var b strings.Builder
	b.WriteString("Posts\n\n")
	if len(m.posts) == 0 {
		b.WriteString("(no posts)\n")
	} else {
		for i, p := range m.posts {
			prefix := "  "
			if i == m.cursor {
				prefix = "> "
			}
			line := prefix + p.title
			if m.width > 0 {
				line = wrapText(line, m.width)
			}
			b.WriteString(line)
			b.WriteString("\n")
		}
	}
	hint := "\n↑/↓ or j/k move • enter open • esc home • q quit\n"
	if m.width > 0 {
		b.WriteString("\n")
		b.WriteString(wrapText(strings.TrimSpace(hint), m.width))
		b.WriteString("\n")
	} else {
		b.WriteString(hint)
	}
	return b.String()
}

func (m *model) viewPost() string {
	if m.cursor < 0 || m.cursor >= len(m.posts) {
		return "Invalid post selected"
	}
	scrollHint := "↑/↓ j/k • PgUp/PgDn • esc list • q quit"
	if m.width > 0 && m.height > 0 {
		return m.vp.View() + "\n\n" + wrapText(scrollHint, m.width)
	}
	p := m.posts[m.cursor]
	return fmt.Sprintf("Post: %s\n\n%s\n\n%s", p.title, p.body, scrollHint)
}

func main() {
	repoRoot := ".."
	postsDir := filepath.Join(repoRoot, "posts")
	posts, err := loadPosts(postsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load posts from %s: %v\n", postsDir, err)
		os.Exit(1)
	}

	homePath := filepath.Join(".", "home.md")
	homeMd, err := loadTUIHome(homePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load home from %s (run from the tui/ directory): %v\n", homePath, err)
		os.Exit(1)
	}

	m0 := &model{
		vp:           viewport.New(0, 0),
		screen:       screenHome,
		posts:        posts,
		homeMarkdown: homeMd,
	}
	p := tea.NewProgram(m0, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v", err)
		os.Exit(1)
	}
}