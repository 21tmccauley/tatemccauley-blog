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
	"github.com/mattn/go-runewidth"

	blog "github.com/tatemccauley/tatemccauley-blog"
)

type screen int

const (
	screenIntro screen = iota
	screenMain
	screenPost
)

// tab is one entry in the global nav bar, mirroring the website's header:
// Home · Blog · About · Resume · Now.
type tab int

const (
	tabHome tab = iota
	tabBlog
	tabAbout
	tabResume
	tabNow
	tabCount
)

var tabTitles = [tabCount]string{"Home", "Blog", "About", "Resume", "Now"}

// isPage reports whether the tab shows a standalone markdown page in the
// scrolling viewport (as opposed to Home and Blog, which compose their own
// content).
func (t tab) isPage() bool { return t == tabAbout || t == tabResume || t == tabNow }

const (
	maxColumn = 76 // readable column cap; content is centered in the terminal
	topMargin = 1
	// Lines the chrome reserves around the scrolling viewport:
	// top(1) + nav(2) + gap(1) + gap(1) + footer(2).
	chromeLines = 7
	// Sessions are anonymous, so the client-reported window size is untrusted.
	// A real terminal never approaches these bounds; clamping stops a forged
	// window-change (dimensions can be up to 2^32) from driving a multi-gigabyte
	// padding allocation and OOM-killing the machine.
	maxWidth  = 500
	maxHeight = 300
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
	tab           tab
	posts         []post
	cursor        int
	frame         int            // intro animation frame counter
	homeMarkdown  string         // raw body from tui/home.md
	homeView      string         // homeMarkdown rendered via glamour at contentWidth
	pageMD        map[tab]string // raw markdown for the About/Resume/Now tabs
	pageView      map[tab]string // pageMD rendered via glamour at contentWidth
	pageScroll    map[tab]int    // saved offsets so each page keeps its place
	md            *glamour.TermRenderer
}

func newModel(r *lipgloss.Renderer, posts []post, homeMarkdown string, pages map[tab]string) *model {
	return &model{
		th:           newTheme(r),
		vp:           viewport.New(0, 0),
		screen:       screenIntro,
		posts:        posts,
		homeMarkdown: homeMarkdown,
		pageMD:       pages,
		pageView:     make(map[tab]string, len(pages)),
		pageScroll:   make(map[tab]int, len(pages)),
	}
}

func (m *model) Init() tea.Cmd {
	return introTickCmd()
}

// setSize recomputes layout for a new terminal size: the content column, the
// glamour renderer (wrap width, color profile) and cached page renders, and
// the viewport dimensions.
func (m *model) setSize(w, h int) {
	if w < 0 {
		w = 0
	}
	if h < 0 {
		h = 0
	}
	if w > maxWidth {
		w = maxWidth
	}
	if h > maxHeight {
		h = maxHeight
	}
	cw := w - 6
	if cw > maxColumn {
		cw = maxColumn
	}
	if cw < 20 {
		cw = 20
	}

	// The glamour renders below are the expensive part; they depend only on the
	// content width, so a height-only resize (or repeated identical events)
	// skips them. The renderer and cached views persist from the last width, so
	// this is a no-op only when they already exist.
	widthChanged := cw != m.contentWidth || m.md == nil
	m.width, m.height = w, h
	m.contentWidth = cw

	if widthChanged {
		if r, err := m.th.markdownRenderer(cw); err == nil {
			m.md = r
		}
		m.homeView = m.renderMarkdown(m.homeMarkdown)
		for t, md := range m.pageMD {
			m.pageView[t] = m.renderMarkdown(md)
		}
	}

	m.vp.Width = cw
	m.vp.Height = m.viewportHeight()
	m.reloadViewportAfterResize()
}

func (m *model) viewportHeight() int {
	h := m.height - chromeLines
	if h < 3 {
		h = 3
	}
	return h
}

// renderMarkdown renders through glamour, falling back to plain wrapped text
// if the renderer is unavailable.
func (m *model) renderMarkdown(s string) string {
	if m.md != nil {
		if out, err := m.md.Render(s); err == nil {
			return strings.TrimRight(out, "\n")
		}
	}
	return wrapText(s, m.contentWidth)
}

func (m *model) renderPostBody() string {
	if m.cursor < 0 || m.cursor >= len(m.posts) {
		return ""
	}
	return m.renderMarkdown(m.posts[m.cursor].body)
}

func (m *model) loadPostIntoViewport() {
	if m.cursor < 0 || m.cursor >= len(m.posts) {
		return
	}
	m.vp.Width = m.contentWidth
	m.vp.Height = m.viewportHeight()
	m.vp.SetContent(m.renderPostBody())
	m.vp.GotoTop()
}

func (m *model) reloadViewportAfterResize() {
	prev := m.vp.YOffset
	switch {
	case m.screen == screenPost:
		if m.cursor < 0 || m.cursor >= len(m.posts) {
			return
		}
		m.vp.SetContent(m.renderPostBody())
	case m.screen == screenMain && m.tab.isPage():
		m.vp.SetContent(m.pageView[m.tab])
	default:
		return
	}
	m.vp.SetYOffset(prev)
}

// switchTab moves the nav selection from any screen, saving and restoring
// page scroll positions so each tab keeps its place.
func (m *model) switchTab(t tab) {
	if t < 0 || t >= tabCount || (t == m.tab && m.screen == screenMain) {
		return
	}
	if m.screen == screenMain && m.tab.isPage() {
		m.pageScroll[m.tab] = m.vp.YOffset
	}
	m.screen = screenMain
	m.tab = t
	if t.isPage() {
		m.vp.Width = m.contentWidth
		m.vp.Height = m.viewportHeight()
		m.vp.SetContent(m.pageView[t])
		m.vp.SetYOffset(m.pageScroll[t])
	}
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
			m.screen = screenMain
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
			m.screen = screenMain // any other key skips the intro
			return m, nil
		case screenMain:
			return m.updateMain(msg)
		case screenPost:
			return m.updatePost(msg)
		}
		return m, nil
	}
	return m, nil
}

// handleNavKey applies the tab-switching keys shared by every screen with a
// nav bar. It reports whether the key was consumed.
func (m *model) handleNavKey(key string) bool {
	switch key {
	case "right", "tab":
		m.switchTab((m.tab + 1) % tabCount)
	case "left", "shift+tab":
		m.switchTab((m.tab + tabCount - 1) % tabCount)
	case "1", "2", "3", "4", "5":
		m.switchTab(tab(key[0] - '1'))
	default:
		return false
	}
	return true
}

func (m *model) updateMain(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "esc":
		m.switchTab(tabHome)
		return m, nil
	}
	if m.handleNavKey(key) {
		return m, nil
	}

	switch m.tab {
	case tabHome:
		if key == "enter" || key == "p" {
			m.switchTab(tabBlog)
		}
	case tabBlog:
		switch key {
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
		}
	default: // page tabs scroll in the viewport
		var cmd tea.Cmd
		m.vp, cmd = m.vp.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m *model) updatePost(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "esc":
		m.screen = screenMain
		return m, nil
	}
	if m.handleNavKey(key) {
		return m, nil
	}
	var cmd tea.Cmd
	m.vp, cmd = m.vp.Update(msg)
	return m, cmd
}

func (m *model) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}
	switch m.screen {
	case screenIntro:
		return m.viewIntro()
	case screenPost:
		return m.viewPost()
	default:
		return m.viewMain()
	}
}

// center places a left-aligned block in the middle of the terminal width.
func (m *model) center(s string) string {
	return m.th.r.PlaceHorizontal(m.width, lipgloss.Center, s)
}

// navBar mirrors the website's global header: brand on the left, page tabs on
// the right with the active one highlighted. On terminals too narrow for
// both, the brand is dropped so the tabs stay on one line.
func (m *model) navBar() string {
	cw := m.contentWidth
	items := make([]string, 0, tabCount)
	for i, title := range tabTitles {
		if tab(i) == m.tab {
			items = append(items, m.th.tabActive.Render(title))
		} else {
			items = append(items, m.th.tabInactive.Render(title))
		}
	}
	tabs := strings.Join(items, "  ")
	line := tabs
	brand := m.th.brand.Render("tatemccauley.com")
	if gap := cw - lipgloss.Width(brand) - lipgloss.Width(tabs); gap >= 2 {
		line = brand + strings.Repeat(" ", gap) + tabs
	}
	return line + "\n" + m.th.rule(cw)
}

// footerBar is the rule + key hints shown at the bottom of every screen.
func (m *model) footerBar(hints string) string {
	return m.th.rule(m.contentWidth) + "\n" + hints
}

func (m *model) viewMain() string {
	var content, hints string
	switch m.tab {
	case tabHome:
		content = m.homeContent()
		hints = m.th.hints(
			[2]string{"←/→", "pages"},
			[2]string{"enter", "posts"},
			[2]string{"q", "quit"},
		)
	case tabBlog:
		content = m.blogContent()
		hints = m.th.hints(
			[2]string{"↑/↓", "move"},
			[2]string{"enter", "open"},
			[2]string{"←/→", "pages"},
			[2]string{"q", "quit"},
		)
	default:
		content = m.vp.View()
		hints = m.th.hints(
			[2]string{"↑/↓", "scroll"},
			[2]string{"←/→", "pages"},
			[2]string{"q", "quit"},
		)
		if pct := scrollLabel(m.vp); pct != "" {
			hints += "   " + m.th.muted.Render(pct)
		}
	}

	inner := lipgloss.JoinVertical(lipgloss.Left,
		m.navBar(),
		"",
		content,
		"",
		m.footerBar(hints),
	)
	return m.center(padTop(inner, topMargin))
}

func (m *model) homeContent() string {
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
			recent.WriteString(m.postRow(m.posts[i], i == 0))
			if i < n-1 {
				recent.WriteByte('\n')
			}
		}
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		splash,
		"",
		m.homeView,
		"",
		recent.String(),
	)
}

func (m *model) blogContent() string {
	if len(m.posts) == 0 {
		return m.th.muted.Render("(no posts yet)")
	}
	var b strings.Builder
	for i, p := range m.posts {
		b.WriteString(m.postRow(p, i == m.cursor))
		if i < len(m.posts)-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

// postRow renders one list row, truncating the title so the row never grows
// past the content column (an overwide row would break centering).
func (m *model) postRow(p post, selected bool) string {
	avail := m.contentWidth - 14 // marker(2) + date(10) + gap(2)
	if avail < 8 {
		avail = 8
	}
	title := runewidth.Truncate(p.title, avail, "…")
	marker, styled := "  ", m.th.itemTitle.Render(title)
	if selected {
		marker = m.th.selMarker.Render("▸ ")
		styled = m.th.selTitle.Render(title)
	}
	return marker + m.th.itemDate.Render(dateCol(p)) + "  " + styled
}

func (m *model) viewPost() string {
	if m.cursor < 0 || m.cursor >= len(m.posts) {
		return m.center("Invalid post")
	}
	hints := m.th.hints(
		[2]string{"↑/↓", "scroll"},
		[2]string{"esc", "back"},
		[2]string{"←/→", "pages"},
		[2]string{"q", "quit"},
	)
	if pct := scrollLabel(m.vp); pct != "" {
		hints += "   " + m.th.muted.Render(pct)
	}

	inner := lipgloss.JoinVertical(lipgloss.Left,
		m.navBar(),
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
	home, err := loadPage(blog.Content, "tui/home.md")
	if err != nil {
		fmt.Fprintf(os.Stderr, "load home: %v\n", err)
		os.Exit(1)
	}
	pages := make(map[tab]string, 3)
	for t, name := range map[tab]string{tabAbout: "about.md", tabResume: "resume.md", tabNow: "now.md"} {
		md, err := loadPage(blog.Content, name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "load %s: %v\n", name, err)
			os.Exit(1)
		}
		pages[t] = md
	}

	if *serve {
		if *httpPort > 0 {
			go serveStatic(fmt.Sprintf("%s:%d", *host, *httpPort), *siteDir)
		}
		runServer(*host, *port, *keyPath, posts, home, pages)
		return
	}

	p := tea.NewProgram(newModel(lipgloss.DefaultRenderer(), posts, home, pages), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v", err)
		os.Exit(1)
	}
}
