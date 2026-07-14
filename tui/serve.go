package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"
)

const (
	// idleTimeout disconnects readers who walk away; sessions are cheap but
	// abandoned PTYs shouldn't pile up on a nano instance.
	idleTimeout = 30 * time.Minute
	maxSessions = 100
)

// runServer serves the blog TUI over SSH. Connections are anonymous — no
// authentication — and each session only ever talks to a fresh Bubble Tea
// model: there is no shell, exec, or filesystem behind the connection.
func runServer(host string, port int, hostKeyPath string, posts []post, homeMD string, pages map[tab]string) {
	teaHandler := func(s ssh.Session) (tea.Model, []tea.ProgramOption) {
		pty, _, _ := s.Pty()
		// Bind styles to the client's session so colors reflect the visitor's
		// terminal, not the server's stdout.
		m := newModel(bubbletea.MakeRenderer(s), posts, homeMD, pages)
		m.setSize(pty.Window.Width, pty.Window.Height)
		return m, []tea.ProgramOption{tea.WithAltScreen()}
	}

	srv, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, strconv.Itoa(port))),
		wish.WithHostKeyPath(hostKeyPath),
		wish.WithIdleTimeout(idleTimeout),
		// Middleware runs last-to-first: logging wraps everything, then the
		// session cap, then the PTY check, with the app innermost.
		wish.WithMiddleware(
			bubbletea.Middleware(teaHandler),
			activeterm.Middleware(),
			limitSessions(maxSessions),
			logging.Middleware(),
		),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "create server: %v\n", err)
		os.Exit(1)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	log.Info("Serving blog over SSH", "host", host, "port", port)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			log.Error("Server error", "error", err)
			done <- syscall.SIGTERM
		}
	}()

	<-done
	log.Info("Shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		fmt.Fprintf(os.Stderr, "shutdown: %v\n", err)
		os.Exit(1)
	}
}

// limitSessions caps concurrent sessions so a connection flood degrades into
// polite rejections instead of resource exhaustion.
func limitSessions(max int) wish.Middleware {
	slots := make(chan struct{}, max)
	return func(next ssh.Handler) ssh.Handler {
		return func(s ssh.Session) {
			select {
			case slots <- struct{}{}:
				defer func() { <-slots }()
				next(s)
			default:
				wish.Println(s, "The blog is busy right now — please try again in a minute.")
			}
		}
	}
}
