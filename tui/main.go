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

const (
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
	padX          int // card padding at the current width
	hintLines     int // rows the footer needs; it wraps on a narrow card
	screen        screen
	tab           tab
	posts         []post
	cursor        int
	frame         int            // intro animation frame counter
	homeMarkdown  string         // raw body from tui/home.md
	homeView      string         // homeMarkdown rendered via glamour at textWidth
	pageMD        map[tab]string // raw markdown for the About/Resume/Now tabs
	pageView      map[tab]string // pageMD rendered via glamour at textWidth
	tabScroll     map[tab]int    // saved offsets so each tab keeps its place
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
		tabScroll:    make(map[tab]int, tabCount),
	}
}

func (m *model) Init() tea.Cmd {
	return introTickCmd()
}

// setSize recomputes layout for a new terminal size: the card geometry, the
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
	cw, padX := cardLayout(w)

	// The glamour renders below are the expensive part; they depend only on the
	// content width, so a height-only resize (or repeated identical events)
	// skips them. The renderer and cached views persist from the last width, so
	// this is a no-op only when they already exist.
	widthChanged := cw != m.contentWidth || m.md == nil
	m.width, m.height = w, h
	m.contentWidth, m.padX = cw, padX
	m.measureChrome()

	if widthChanged {
		if r, err := m.th.markdownRenderer(m.textWidth()); err == nil {
			m.md = r
		}
		m.homeView = m.renderMarkdown(m.homeMarkdown)
		for t, md := range m.pageMD {
			m.pageView[t] = m.renderMarkdown(md)
		}
	}
	m.syncViewport(m.vp.YOffset)
}

// bodyHeight sizes the viewport to the content it is about to hold, capped by
// what the terminal leaves once the chrome is paid for. This is what makes the
// card hug a short screen instead of stretching to the bottom of the pane, and
// grow to fill it on a long post. Content never depends on this height, so
// there is no circularity in measuring it first.
func (m *model) bodyHeight(body string) int {
	avail := m.height - 2*cardMargY - m.chromeHeight()
	if avail < minBody {
		avail = minBody
	}
	h := lipgloss.Height(body)
	if h > avail {
		h = avail
	}
	if h < 1 {
		h = 1
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
	return wrapText(s, m.textWidth())
}

func (m *model) renderPostBody() string {
	if m.cursor < 0 || m.cursor >= len(m.posts) {
		return ""
	}
	return m.renderMarkdown(m.posts[m.cursor].body)
}

// bodyFor returns the content for the current screen and tab. Every view goes
// through the viewport, so the card sizes itself to its content and can never
// outgrow the terminal however much content a screen holds.
func (m *model) bodyFor() string {
	return m.padBlock(m.rawBodyFor())
}

func (m *model) rawBodyFor() string {
	if m.screen == screenPost {
		return m.renderPostBody()
	}
	switch m.tab {
	case tabHome:
		return m.homeContent()
	case tabBlog:
		return m.blogContent()
	default:
		return m.pageView[m.tab]
	}
}

// padBlock pads every line of a body out to the text column with the card
// background. Two things downstream pad short lines with *unstyled* space,
// which shows the visitor's terminal through the middle of the card:
// viewport.View renders through a background-less lipgloss style, and
// JoinVertical pads with bare strings.Repeat(" "). Glamour output already
// arrives full width, so this is a no-op there and matters for the rows this
// package composes itself — list rows, section headings, the splash.
func (m *model) padBlock(s string) string {
	tw := m.textWidth()
	lines := strings.Split(s, "\n")
	for i, ln := range lines {
		lines[i] = ln + m.th.pad(tw-lipgloss.Width(ln))
	}
	return strings.Join(lines, "\n")
}

// enterMain leaves the boot screen for the home tab, filling the viewport the
// intro has no use for.
func (m *model) enterMain() {
	m.screen = screenMain
	m.syncViewport(0)
}

// syncViewport reloads the viewport for the current view and scrolls it to off.
// The offset is applied after the new content, because viewport.SetYOffset
// clamps against whatever it currently holds: setting it first would measure a
// restored position against the view being left, and silently truncate it to
// zero whenever that view happened to fit on one screen.
func (m *model) syncViewport(off int) {
	body := m.bodyFor()
	m.vp.Width = m.textWidth()
	m.vp.Height = m.bodyHeight(body)
	m.vp.SetContent(body)
	m.vp.SetYOffset(off)
	if m.screen == screenMain && m.tab == tabBlog {
		m.revealCursor()
	}
}

// revealCursor scrolls the post list just far enough to keep the selected row
// on screen.
func (m *model) revealCursor() {
	row := m.cursor
	switch {
	case row < m.vp.YOffset:
		m.vp.SetYOffset(row)
	case row >= m.vp.YOffset+m.vp.Height:
		m.vp.SetYOffset(row - m.vp.Height + 1)
	}
}

func (m *model) openPost() {
	if m.cursor < 0 || m.cursor >= len(m.posts) {
		return
	}
	m.tabScroll[tabBlog] = m.vp.YOffset // so esc returns the list as it was
	m.screen = screenPost
	m.syncViewport(0)
}

// switchTab moves the nav selection from any screen, saving and restoring
// scroll positions so each tab keeps its place.
func (m *model) switchTab(t tab) {
	if t < 0 || t >= tabCount || (t == m.tab && m.screen == screenMain) {
		return
	}
	if m.screen == screenMain {
		m.tabScroll[m.tab] = m.vp.YOffset
	}
	m.screen = screenMain
	m.tab = t
	m.syncViewport(m.tabScroll[t])
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
			m.enterMain()
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
			m.enterMain() // any other key skips the intro
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
	case tabBlog:
		// The list is driven entirely by the cursor, and every key that could
		// move the view has to move the cursor with it. Handing leftovers to the
		// viewport instead would let its own pgdn/space/d bindings scroll the
		// selection off screen, and then enter opens a post the reader cannot
		// see. So the list consumes every key it does not act on.
		switch key {
		case "up", "k":
			m.moveCursor(-1)
		case "down", "j":
			m.moveCursor(1)
		case "pgup", "b":
			m.moveCursor(-m.vp.Height)
		case "pgdown", "f", " ":
			m.moveCursor(m.vp.Height)
		case "home", "g":
			m.moveCursor(-len(m.posts))
		case "end", "G":
			m.moveCursor(len(m.posts))
		case "enter":
			if len(m.posts) > 0 {
				m.openPost()
			}
		}
		return m, nil
	case tabHome:
		if key == "enter" || key == "p" {
			m.switchTab(tabBlog)
			return m, nil
		}
	}

	// Anything left over scrolls the viewport.
	var cmd tea.Cmd
	m.vp, cmd = m.vp.Update(msg)
	return m, cmd
}

// moveCursor shifts the post-list selection by delta, clamped to the list, and
// scrolls just enough to keep it visible.
func (m *model) moveCursor(delta int) {
	if len(m.posts) == 0 {
		return
	}
	next := m.cursor + delta
	if next < 0 {
		next = 0
	}
	if next > len(m.posts)-1 {
		next = len(m.posts) - 1
	}
	if next == m.cursor {
		return
	}
	m.cursor = next
	m.syncViewport(m.vp.YOffset)
}

func (m *model) updatePost(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "esc":
		m.screen = screenMain
		m.syncViewport(m.tabScroll[m.tab])
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

func (m *model) viewMain() string {
	hints := pageHints
	switch m.tab {
	case tabHome:
		hints = homeHints
	case tabBlog:
		hints = blogHints
	}
	return m.card(m.scrollbar(m.vp.View()), m.th.hints(hints...))
}

func (m *model) homeContent() string {
	tw := m.textWidth()

	// No bordered splash here: the card is already a frame, and a box inside a
	// box reads as clutter. The wordmark carries it on its own.
	splash := m.th.center.Width(tw).Render(
		m.th.brand.Render(letterSpace("TATE McCAULEY")) + "\n" +
			m.th.tagline.Render("the terminal edition"))

	var recent strings.Builder
	recent.WriteString(m.th.section.Render("RECENT"))
	recent.WriteByte('\n')
	recent.WriteString(m.th.pad(tw))
	recent.WriteByte('\n')
	if len(m.posts) == 0 {
		recent.WriteString(m.th.muted.Render("(no posts yet)"))
	} else {
		n := 3
		if len(m.posts) < n {
			n = len(m.posts)
		}
		for i := 0; i < n; i++ {
			recent.WriteString(m.postRow(m.posts[i], false))
			if i < n-1 {
				recent.WriteByte('\n')
			}
		}
	}

	blank := m.th.pad(tw)
	return lipgloss.JoinVertical(lipgloss.Left,
		splash,
		blank,
		m.homeView,
		blank,
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
// past the text column (an overwide row would break the card).
func (m *model) postRow(p post, selected bool) string {
	avail := m.textWidth() - 14 // marker(2) + date(10) + gap(2)
	if avail < 8 {
		avail = 8
	}
	title := runewidth.Truncate(p.title, avail, "…")
	marker, styled := m.th.pad(2), m.th.itemTitle.Render(title)
	if selected {
		marker = m.th.selMarker.Render("› ")
		styled = m.th.selTitle.Render(title)
	}
	return marker + m.th.itemDate.Render(dateCol(p)) + m.th.pad(2) + styled
}

func (m *model) viewPost() string {
	if m.cursor < 0 || m.cursor >= len(m.posts) {
		return m.card(m.th.muted.Render("Invalid post"), m.th.hints(postHints...))
	}
	title := runewidth.Truncate(m.posts[m.cursor].title, m.textWidth(), "…")
	body := lipgloss.JoinVertical(lipgloss.Left,
		m.fillLine(m.th.section.Render(title)),
		m.th.pad(m.contentWidth),
		m.scrollbar(m.vp.View()),
	)
	return m.card(body, m.th.hints(postHints...))
}

// dateCol formats a post date as a fixed-width column so titles align; a
// missing date becomes blank padding of the same width.
func dateCol(p post) string {
	if p.date.IsZero() {
		return "          " // 10 spaces == len("2006-01-02")
	}
	return p.date.Format("2006-01-02")
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
