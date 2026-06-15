package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"claude-stats/internal/models"
)

func dirSize(path string) int64 {
	var n int64
	filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			n += info.Size()
		}
		return nil
	})
	return n
}

func projectDisplayName(dirName string) string {
	decoded := strings.ReplaceAll(dirName, "-", "/")
	parts := strings.Split(decoded, "/")
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] != "" {
			return parts[i]
		}
	}
	return dirName
}

func Analyze(claudeDir string) models.StorageInfo {
	byDir := make(map[string]int64)
	var total int64

	for _, sub := range []string{"projects", "file-history", "plugins", "backups", "cache",
		"telemetry", "shell-snapshots", "sessions", "ide", "session-env"} {
		if n := dirSize(filepath.Join(claudeDir, sub)); n > 0 {
			byDir[sub] = n
			total += n
		}
	}
	for _, f := range []string{"stats-cache.json", "history.jsonl", "settings.json", "mcp-needs-auth-cache.json"} {
		if info, err := os.Stat(filepath.Join(claudeDir, f)); err == nil {
			byDir[f] = info.Size()
			total += info.Size()
		}
	}

	var projects []models.ProjectStorage
	projectsDir := filepath.Join(claudeDir, "projects")
	if dirs, err := os.ReadDir(projectsDir); err == nil {
		for _, d := range dirs {
			if !d.IsDir() {
				continue
			}
			n := dirSize(filepath.Join(projectsDir, d.Name()))
			sessions := 0
			if files, err := os.ReadDir(filepath.Join(projectsDir, d.Name())); err == nil {
				for _, f := range files {
					if strings.HasSuffix(f.Name(), ".jsonl") {
						sessions++
					}
				}
			}
			projects = append(projects, models.ProjectStorage{
				Project:  projectDisplayName(d.Name()),
				Bytes:    n,
				Sessions: sessions,
			})
		}
	}
	sort.Slice(projects, func(i, j int) bool { return projects[i].Bytes > projects[j].Bytes })
	if len(projects) > 20 {
		projects = projects[:20]
	}

	return models.StorageInfo{TotalBytes: total, ByDir: byDir, ByProject: projects}
}

type Settings struct {
	Raw     map[string]interface{}
	MCP     map[string]interface{}
	Active  []map[string]interface{}
	History []HistoryEntry
}

type HistoryEntry struct {
	Display   string `json:"display"`
	Timestamp int64  `json:"timestamp"`
	Project   string `json:"project"`
	SessionId string `json:"sessionId"`
}

func ReadSettings(claudeDir string) Settings {
	s := Settings{
		Raw: make(map[string]interface{}),
		MCP: make(map[string]interface{}),
	}
	if data, err := os.ReadFile(filepath.Join(claudeDir, "settings.json")); err == nil {
		json.Unmarshal(data, &s.Raw)
	}
	if data, err := os.ReadFile(filepath.Join(claudeDir, "mcp-needs-auth-cache.json")); err == nil {
		json.Unmarshal(data, &s.MCP)
	}
	sessionsDir := filepath.Join(claudeDir, "sessions")
	if files, err := os.ReadDir(sessionsDir); err == nil {
		for _, f := range files {
			if strings.HasSuffix(f.Name(), ".json") {
				if data, err := os.ReadFile(filepath.Join(sessionsDir, f.Name())); err == nil {
					var m map[string]interface{}
					if json.Unmarshal(data, &m) == nil {
						s.Active = append(s.Active, m)
					}
				}
			}
		}
	}
	return s
}
