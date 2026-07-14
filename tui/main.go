package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	blog "github.com/tatemccauley/tatemccauley-blog"
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
	date  time.Time
	body  string
}

type model struct {
	vp            viewport.Model
	width, height int
	screen        screen
	posts         []post
	cursor        int
	homeMarkdown  string // body from tui/home.md (optional front matter stripped)
}

func newModel(posts []post, homeMarkdown string) *model {
	return &model{
		vp:           viewport.New(0, 0),
		screen:       screenHome,
		posts:        posts,
		homeMarkdown: homeMarkdown,
	}
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
			if !p.date.IsZero() {
				line = prefix + p.date.Format("2006-01-02") + "  " + p.title
			}
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
	var (
		serve    = flag.Bool("serve", false, "serve the TUI over SSH instead of running it locally")
		host     = flag.String("host", "0.0.0.0", "address to listen on (with -serve)")
		port     = flag.Int("port", 23234, "SSH port to listen on (with -serve)")
		keyPath  = flag.String("host-key", ".ssh/blog_host_ed25519", "SSH host key path, created on first run (with -serve)")
		httpPort = flag.Int("http-port", 0, "also serve the built website over HTTP on this port; 0 disables (with -serve)")
		siteDir  = flag.String("site-dir", "_site", "directory of the built Eleventy site to serve (with -http-port)")
	)
	flag.Parse()

	posts, err := loadPosts(blog.Content)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load posts: %v\n", err)
		os.Exit(1)
	}
	home, err := loadHome(blog.Content)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load home: %v\n", err)
		os.Exit(1)
	}

	if *serve {
		if *httpPort > 0 {
			go serveStatic(fmt.Sprintf("%s:%d", *host, *httpPort), *siteDir)
		}
		runServer(*host, *port, *keyPath, posts, home)
		return
	}

	p := tea.NewProgram(newModel(posts, home), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v", err)
		os.Exit(1)
	}
}
