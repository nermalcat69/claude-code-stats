package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"claude-stats/embedded"
	"claude-stats/internal/cache"
	"claude-stats/internal/handlers"
	"claude-stats/tui"
)

const (
	defaultPort     = 6967
	refreshInterval = 5 * time.Minute
)

func claudeDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".claude")
}

func isTTY() bool {
	fi, err := os.Stdout.Stat()
	return err == nil && (fi.Mode()&os.ModeCharDevice) != 0
}

func main() {
	port := defaultPort
	if p := os.Getenv("PORT"); p != "" {
		fmt.Sscanf(p, "%d", &port)
	}

	dir := claudeDir()
	log.SetFlags(log.Ltime)

	// ── Cache ─────────────────────────────────────────────────────────────────
	dc := cache.New(dir)
	go func() {
		dc.Refresh()
		dc.StartAutoRefresh(refreshInterval)
	}()

	// ── HTTP server ───────────────────────────────────────────────────────────
	mux := http.NewServeMux()
	srv := handlers.New(dc, dir)
	srv.RegisterRoutes(mux)

	mux.Handle("/", http.FileServer(http.FS(embedded.FS())))

	go func() {
		addr := fmt.Sprintf(":%d", port)
		if err := http.ListenAndServe(addr, srv.WithCORS(mux)); err != nil {
			log.Fatalf("http: %v", err)
		}
	}()

	// ── TUI or headless ───────────────────────────────────────────────────────
	if isTTY() {
		m := tui.New(dc, port)
		p := tea.NewProgram(m, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "tui: %v\n", err)
		}
	} else {
		// Headless / daemon mode
		log.Printf("claude-stats: dashboard at http://localhost:%d", port)
		log.Printf("claude-stats: reading ~/.claude from %s", dir)
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		log.Println("claude-stats: shutting down")
	}
}
