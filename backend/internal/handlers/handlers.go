package handlers

import (
	"bufio"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"claude-stats/internal/cache"
	"claude-stats/internal/models"
	"claude-stats/internal/pricing"
	"claude-stats/internal/storage"
)

type Server struct {
	cache     *cache.DataCache
	claudeDir string
}

func New(c *cache.DataCache, claudeDir string) *Server {
	return &Server{cache: c, claudeDir: claudeDir}
}

func (s *Server) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/overview", s.Overview)
	mux.HandleFunc("/api/projects", s.Projects)
	mux.HandleFunc("/api/sessions", s.Sessions)
	mux.HandleFunc("/api/history", s.History)
	mux.HandleFunc("/api/settings", s.Settings)
	mux.HandleFunc("/api/storage", s.Storage)
	mux.HandleFunc("/api/refresh", s.Refresh)
	mux.HandleFunc("/api/status", s.Status)
}

func (s *Server) WithCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func jsonOK(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func (s *Server) Overview(w http.ResponseWriter, r *http.Request) {
	stats, _, _, derived := s.cache.Get()
	if stats == nil {
		stats = &models.StatsCache{}
	}
	costs := make(map[string]float64)
	var total float64
	for model, u := range stats.ModelUsage {
		c := pricing.Compute(model, u.InputTokens, u.OutputTokens,
			u.CacheReadInputTokens, u.CacheCreationInputTokens)
		costs[model] = c
		total += c
	}
	jsonOK(w, models.OverviewResponse{
		Stats:         stats,
		ComputedCosts: costs,
		TotalCostUSD:  total,
		ThinkingWPM:   derived.ThinkingWPM,
	})
}

func (s *Server) Projects(w http.ResponseWriter, r *http.Request) {
	_, _, projects, _ := s.cache.Get()
	jsonOK(w, projects)
}

func (s *Server) Sessions(w http.ResponseWriter, r *http.Request) {
	_, sessions, _, _ := s.cache.Get()
	jsonOK(w, sessions)
}

func (s *Server) History(w http.ResponseWriter, r *http.Request) {
	type Entry struct {
		Display   string `json:"display"`
		Timestamp int64  `json:"timestamp"`
		Project   string `json:"project"`
	}
	type CmdFreq struct {
		Command string `json:"command"`
		Count   int    `json:"count"`
	}

	var entries []Entry
	f, err := os.Open(filepath.Join(s.claudeDir, "history.jsonl"))
	if err == nil {
		defer f.Close()
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			var e Entry
			if json.Unmarshal(sc.Bytes(), &e) == nil {
				entries = append(entries, e)
			}
		}
	}

	counts := make(map[string]int)
	for _, e := range entries {
		if strings.HasPrefix(e.Display, "/") {
			counts[e.Display]++
		}
	}
	var freq []CmdFreq
	for cmd, n := range counts {
		freq = append(freq, CmdFreq{cmd, n})
	}
	sort.Slice(freq, func(i, j int) bool { return freq[i].Count > freq[j].Count })
	if len(freq) > 15 {
		freq = freq[:15]
	}

	jsonOK(w, map[string]interface{}{
		"totalEntries":  len(entries),
		"commandCounts": freq,
	})
}

func (s *Server) Settings(w http.ResponseWriter, r *http.Request) {
	cfg := storage.ReadSettings(s.claudeDir)
	jsonOK(w, map[string]interface{}{
		"settings":       cfg.Raw,
		"mcp":            cfg.MCP,
		"activeSessions": cfg.Active,
	})
}

func (s *Server) Storage(w http.ResponseWriter, r *http.Request) {
	jsonOK(w, storage.Analyze(s.claudeDir))
}

func (s *Server) Refresh(w http.ResponseWriter, r *http.Request) {
	go s.cache.Refresh()
	jsonOK(w, map[string]string{"status": "refreshing"})
}

func (s *Server) Status(w http.ResponseWriter, r *http.Request) {
	jsonOK(w, map[string]interface{}{
		"status":      "ok",
		"ready":       s.cache.IsReady(),
		"lastRefresh": s.cache.LastRefresh(),
	})
}
