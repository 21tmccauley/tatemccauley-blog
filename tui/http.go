package main

import (
	"net/http"
	"time"

	"github.com/charmbracelet/log"
)

// serveStatic serves the pre-built Eleventy site (dir) over plain HTTP. TLS is
// terminated at the Fly edge, so this only ever speaks HTTP on an internal
// port. It runs alongside the SSH server (see runServer) so one machine serves
// both the browser site and the terminal edition from a single IP.
func serveStatic(addr, dir string) {
	srv := &http.Server{
		Addr:              addr,
		Handler:           http.FileServer(http.Dir(dir)),
		ReadHeaderTimeout: 10 * time.Second,
	}
	log.Info("Serving website over HTTP", "addr", addr, "dir", dir)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		// The site being down shouldn't take SSH with it; Fly's health check
		// on the HTTP port will restart the machine if this is fatal.
		log.Error("HTTP server error", "error", err)
	}
}
