package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"

	blog "github.com/tatemccauley/tatemccauley-blog"
)

type screen int

const (
	screenIntro screen = iota
	screenHome
	screenList
	screenPost
)

const (
	maxColumn = 76 // readable column cap; content is centered in the terminal
	topMargin = 1
	// Lines the post chrome reserves around the scrolling viewport:
	// top(1) + header(2) + gap(1) + gap(1) + footer(2).
	postChromeLines = 7
)

type post struct {
	title string
	date  time.Time
	body  string
}

type model struct {
	th            *theme
	vp            viewport.Model
	width, height int
	contentWidth  int
	screen        screen
	posts         []post
	cursor        int
	frame         int    // intro animation frame counter
	homeMarkdown  string // raw body from tui/home.md
	homeView      string // homeMarkdown rendered via glamour at contentWidth
	md            *glamour.TermRenderer
}

func newModel(r *lipgloss.Renderer, posts []post, homeMarkdown string) *model {
	return &model{
		th:           newTheme(r),
		vp:           viewport.New(0, 0),
		screen:       screenIntro,
		posts:        posts,
		homeMarkdown: homeMarkdown,
	}
}

func (m *model) Init() tea.Cmd {
	return introTickCmd()
}

// setSize recomputes layout for a new terminal size: the content column, the
// glamour renderer (wrap width, color profile) and cached home render, and the
// post viewport dimensions.
func (m *model) setSize(w, h int) {
	m.width, m.height = w, h
	cw := w - 6
	if cw > maxColumn {
		cw = maxColumn
	}
	if cw < 20 {
		cw = 20
	}
	m.contentWidth = cw

	if r, err := m.th.markdownRenderer(cw); err == nil {
		m.md = r
		if out, err := r.Render(m.homeMarkdown); err == nil {
			m.homeView = strings.TrimRight(out, "\n")
		} else {
			m.homeView = m.homeMarkdown
		}
	}

	m.vp.Width = cw
	m.vp.Height = m.postBodyHeight()
	m.reloadPostViewportAfterResize()
}

func (m *model) postBodyHeight() int {
	h := m.height - postChromeLines
	if h < 3 {
		h = 3
	}
	return h
}

// renderPostBody renders the selected post's markdown through glamour, falling
// back to plain wrapped text if the renderer is unavailable.
func (m *model) renderPostBody() string {
	if m.cursor < 0 || m.cursor >= len(m.posts) {
		return ""
	}
	body := m.posts[m.cursor].body
	if m.md != nil {
		if out, err := m.md.Render(body); err == nil {
			return strings.TrimRight(out, "\n")
		}
	}
	return wrapText(body, m.contentWidth)
}

func (m *model) loadPostIntoViewport() {
	if m.cursor < 0 || m.cursor >= len(m.posts) {
		return
	}
	m.vp.Width = m.contentWidth
	m.vp.Height = m.postBodyHeight()
	m.vp.SetContent(m.renderPostBody())
	m.vp.GotoTop()
}

func (m *model) reloadPostViewportAfterResize() {
	if m.screen != screenPost || m.cursor < 0 || m.cursor >= len(m.posts) {
		return
	}
	prev := m.vp.YOffset
	m.vp.SetContent(m.renderPostBody())
	m.vp.SetYOffset(prev)
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.setSize(msg.Width, msg.Height)
		return m, nil
	case introTickMsg:
		if m.screen != screenIntro {
			return m, nil // skipped already; stop the tick loop
		}
		m.frame++
		if m.frame > introFrames {
			m.screen = screenHome
			return m, nil
		}
		return m, introTickCmd()
	case tea.KeyMsg:
		switch m.screen {
		case screenIntro:
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			}
			m.screen = screenHome // any other key skips the intro
			return m, nil
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
	if m.width == 0 || m.height == 0 {
		return ""
	}
	switch m.screen {
	case screenIntro:
		return m.viewIntro()
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

// center places a left-aligned block in the middle of the terminal width.
func (m *model) center(s string) string {
	return m.th.r.PlaceHorizontal(m.width, lipgloss.Center, s)
}

// headerBar is the brand line + rule shown on the list and post screens.
func (m *model) headerBar(crumb string) string {
	cw := m.contentWidth
	brand := m.th.brand.Render("tatemccauley.com")
	right := m.th.crumb.Render(crumb)
	gap := cw - lipgloss.Width(brand) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}
	return brand + strings.Repeat(" ", gap) + right + "\n" + m.th.rule(cw)
}

// footerBar is the rule + key hints shown at the bottom of every screen.
func (m *model) footerBar(hints string) string {
	return m.th.rule(m.contentWidth) + "\n" + hints
}

func (m *model) viewHome() string {
	cw := m.contentWidth

	splash := m.th.splashBox.Render(
		m.th.brand.Render("Tate McCauley") + "\n" +
			m.th.tagline.Render("the terminal edition"))
	splash = m.th.center.Width(cw).Render(splash)

	var recent strings.Builder
	recent.WriteString(m.th.section.Render("RECENT"))
	recent.WriteByte('\n')
	if len(m.posts) == 0 {
		recent.WriteString(m.th.muted.Render("(no posts yet)"))
	} else {
		n := 3
		if len(m.posts) < n {
			n = len(m.posts)
		}
		for i := 0; i < n; i++ {
			p := m.posts[i]
			marker, title := "  ", m.th.itemTitle.Render(p.title)
			if i == 0 {
				marker = m.th.selMarker.Render("▸ ")
				title = m.th.selTitle.Render(p.title)
			}
			recent.WriteString(marker + m.th.itemDate.Render(dateCol(p)) + "  " + title)
			if i < n-1 {
				recent.WriteByte('\n')
			}
		}
	}

	inner := lipgloss.JoinVertical(lipgloss.Left,
		splash,
		"",
		m.homeView,
		"",
		recent.String(),
		"",
		m.footerBar(m.th.hints(
			[2]string{"enter", "read posts"},
			[2]string{"q", "quit"},
		)),
	)
	return m.center(padTop(inner, topMargin))
}

func (m *model) viewList() string {
	var b strings.Builder
	if len(m.posts) == 0 {
		b.WriteString(m.th.muted.Render("(no posts yet)"))
	} else {
		for i, p := range m.posts {
			marker, title := "  ", m.th.itemTitle.Render(p.title)
			if i == m.cursor {
				marker = m.th.selMarker.Render("▸ ")
				title = m.th.selTitle.Render(p.title)
			}
			b.WriteString(marker + m.th.itemDate.Render(dateCol(p)) + "  " + title)
			if i < len(m.posts)-1 {
				b.WriteByte('\n')
			}
		}
	}

	inner := lipgloss.JoinVertical(lipgloss.Left,
		m.headerBar("posts"),
		"",
		b.String(),
		"",
		m.footerBar(m.th.hints(
			[2]string{"↑/↓", "move"},
			[2]string{"enter", "open"},
			[2]string{"esc", "home"},
			[2]string{"q", "quit"},
		)),
	)
	return m.center(padTop(inner, topMargin))
}

func (m *model) viewPost() string {
	if m.cursor < 0 || m.cursor >= len(m.posts) {
		return m.center("Invalid post")
	}
	hints := m.th.hints(
		[2]string{"↑/↓", "scroll"},
		[2]string{"esc", "back"},
		[2]string{"q", "quit"},
	)
	if pct := scrollLabel(m.vp); pct != "" {
		hints += "   " + m.th.muted.Render(pct)
	}

	inner := lipgloss.JoinVertical(lipgloss.Left,
		m.headerBar("reading"),
		"",
		m.vp.View(),
		"",
		m.footerBar(hints),
	)
	return m.center(padTop(inner, topMargin))
}

// dateCol formats a post date as a fixed-width column so titles align; a
// missing date becomes blank padding of the same width.
func dateCol(p post) string {
	if p.date.IsZero() {
		return "          " // 10 spaces == len("2006-01-02")
	}
	return p.date.Format("2006-01-02")
}

func scrollLabel(vp viewport.Model) string {
	if vp.AtTop() && vp.AtBottom() {
		return ""
	}
	return fmt.Sprintf("%3.0f%%", vp.ScrollPercent()*100)
}

func padTop(s string, n int) string {
	if n <= 0 {
		return s
	}
	return strings.Repeat("\n", n) + s
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

	p := tea.NewProgram(newModel(lipgloss.DefaultRenderer(), posts, home), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v", err)
		os.Exit(1)
	}
}
