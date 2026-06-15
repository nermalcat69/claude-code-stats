package parsers

import (
	"bufio"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"claude-stats/internal/models"
	"claude-stats/internal/pricing"
)

// rawEntry is the minimal shape of any JSONL line we care about.
type rawEntry struct {
	Type             string          `json:"type"`
	SessionId        string          `json:"sessionId"`
	Timestamp        string          `json:"timestamp"`
	AiTitle          string          `json:"aiTitle,omitempty"`
	LastPrompt       string          `json:"lastPrompt,omitempty"`
	Message          json.RawMessage `json:"message,omitempty"`
	EntryPoint       string          `json:"entrypoint,omitempty"`
	GitBranch        string          `json:"gitBranch,omitempty"`
	Version          string          `json:"version,omitempty"`
	IsSnapshotUpdate bool            `json:"isSnapshotUpdate,omitempty"`
	Snapshot         json.RawMessage `json:"snapshot,omitempty"`
}

type assistantMsg struct {
	Model      string `json:"model"`
	StopReason string `json:"stop_reason"`
	Content    []struct {
		Type     string `json:"type"`
		Name     string `json:"name,omitempty"`
		Thinking string `json:"thinking,omitempty"`
		Text     string `json:"text,omitempty"`
	} `json:"content"`
	Usage struct {
		InputTokens              int64 `json:"input_tokens"`
		OutputTokens             int64 `json:"output_tokens"`
		CacheReadInputTokens     int64 `json:"cache_read_input_tokens"`
		CacheCreationInputTokens int64 `json:"cache_creation_input_tokens"`
	} `json:"usage"`
}

type userMsg struct {
	Content []struct {
		Type    string `json:"type"`
		IsError bool   `json:"is_error,omitempty"`
	} `json:"content"`
}

func parseTime(s string) (time.Time, bool) {
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

// ParseSession reads a single JSONL file and returns a SessionInfo.
func ParseSession(filePath, project string) models.SessionInfo {
	f, err := os.Open(filePath)
	if err != nil {
		return models.SessionInfo{}
	}
	defer f.Close()

	s := models.SessionInfo{
		Project:   project,
		ToolCalls: make(map[string]int),
	}

	var (
		firstTime, lastTime time.Time
		prevUserTime        time.Time
		filesSet            = make(map[string]bool)
	)

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 10*1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		if len(line) < 2 || line[0] != '{' {
			continue
		}

		var e rawEntry
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			continue
		}

		if s.SessionId == "" && e.SessionId != "" {
			s.SessionId = e.SessionId
		}

		ts, hasTS := parseTime(e.Timestamp)
		if hasTS {
			if firstTime.IsZero() || ts.Before(firstTime) {
				firstTime = ts
			}
			if ts.After(lastTime) {
				lastTime = ts
			}
		}

		switch e.Type {
		case "assistant":
			s.MessageCount++
			var msg assistantMsg
			if len(e.Message) > 0 && json.Unmarshal(e.Message, &msg) == nil {
				if s.Model == "" && msg.Model != "" && msg.Model != "<synthetic>" {
					s.Model = msg.Model
				}
				s.TokenUsage.InputTokens += msg.Usage.InputTokens
				s.TokenUsage.OutputTokens += msg.Usage.OutputTokens
				s.TokenUsage.CacheReadInputTokens += msg.Usage.CacheReadInputTokens
				s.TokenUsage.CacheCreationInputTokens += msg.Usage.CacheCreationInputTokens
				s.CostUSD += pricing.Compute(msg.Model,
					msg.Usage.InputTokens, msg.Usage.OutputTokens,
					msg.Usage.CacheReadInputTokens, msg.Usage.CacheCreationInputTokens)

				for _, b := range msg.Content {
					switch b.Type {
					case "tool_use":
						s.ToolCalls[b.Name]++
						s.TotalToolCalls++
					case "thinking":
						s.HasThinking = true
						s.ThinkingChars += int64(len(b.Thinking))
					case "text":
						// text blocks counted but not stored
					}
				}

				// Track API time: time from previous user msg to this assistant msg
				if hasTS && !prevUserTime.IsZero() {
					delta := ts.Sub(prevUserTime).Milliseconds()
					if delta > 0 && delta < 600_000 { // cap at 10 min per turn
						s.ApiTimeMs += delta
					}
					prevUserTime = time.Time{}
				}
			}

		case "user":
			s.MessageCount++
			if e.EntryPoint != "" && s.EntryPoint == "" {
				s.EntryPoint = e.EntryPoint
			}
			if e.GitBranch != "" && s.GitBranch == "" {
				s.GitBranch = e.GitBranch
			}
			if e.Version != "" && s.Version == "" {
				s.Version = e.Version
			}
			if hasTS {
				prevUserTime = ts
			}
			var msg userMsg
			if len(e.Message) > 0 && json.Unmarshal(e.Message, &msg) == nil {
				for _, c := range msg.Content {
					if c.Type == "tool_result" && c.IsError {
						s.ErrorCount++
					}
				}
			}

		case "ai-title":
			if e.AiTitle != "" {
				s.Title = e.AiTitle
			}

		case "last-prompt":
			if e.LastPrompt != "" {
				s.LastPrompt = e.LastPrompt
			}

		case "file-history-snapshot":
			if e.IsSnapshotUpdate && len(e.Snapshot) > 0 {
				var snap struct {
					TrackedFileBackups map[string]json.RawMessage `json:"trackedFileBackups"`
				}
				if json.Unmarshal(e.Snapshot, &snap) == nil {
					for path := range snap.TrackedFileBackups {
						filesSet[path] = true
					}
				}
			}
		}
	}

	if !firstTime.IsZero() {
		s.StartTime = firstTime.Format(time.RFC3339)
		s.EndTime = lastTime.Format(time.RFC3339)
		dur := lastTime.Sub(firstTime)
		if dur > 10*time.Hour {
			dur = 0
		}
		s.DurationMs = dur.Milliseconds()
	}
	s.FileCount = len(filesSet)
	return s
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

func projectDecodedPath(dirName string) string {
	if strings.HasPrefix(dirName, "-") {
		return strings.ReplaceAll(dirName, "-", "/")
	}
	return dirName
}

// ParseAllSessions scans the projects directory and parses all JSONL files concurrently.
func ParseAllSessions(claudeDir string) ([]models.SessionInfo, []models.ProjectInfo) {
	projectsDir := filepath.Join(claudeDir, "projects")

	dirs, err := os.ReadDir(projectsDir)
	if err != nil {
		log.Printf("projects dir: %v", err)
		return nil, nil
	}

	type task struct {
		dirName   string
		files     []string
		totalSize int64
	}

	var tasks []task
	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}
		entries, err := os.ReadDir(filepath.Join(projectsDir, dir.Name()))
		if err != nil {
			continue
		}
		var jsonlFiles []string
		var totalSize int64
		for _, f := range entries {
			if info, err := f.Info(); err == nil {
				totalSize += info.Size()
			}
			if !f.IsDir() && strings.HasSuffix(f.Name(), ".jsonl") {
				jsonlFiles = append(jsonlFiles, filepath.Join(projectsDir, dir.Name(), f.Name()))
			}
		}
		if len(jsonlFiles) > 0 {
			tasks = append(tasks, task{dir.Name(), jsonlFiles, totalSize})
		}
	}

	var (
		allSessions []models.SessionInfo
		allProjects []models.ProjectInfo
		mu          sync.Mutex
		wg          sync.WaitGroup
		sem         = make(chan struct{}, 8)
	)

	for _, t := range tasks {
		name := projectDisplayName(t.dirName)
		path := projectDecodedPath(t.dirName)

		mu.Lock()
		allProjects = append(allProjects, models.ProjectInfo{
			Name:         name,
			Path:         path,
			SessionCount: len(t.files),
			SizeBytes:    t.totalSize,
		})
		mu.Unlock()

		for _, fp := range t.files {
			wg.Add(1)
			go func(fp, pname string) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()
				s := ParseSession(fp, pname)
				mu.Lock()
				allSessions = append(allSessions, s)
				mu.Unlock()
			}(fp, name)
		}
	}
	wg.Wait()

	// Aggregate session stats into projects
	projIdx := make(map[string]int, len(allProjects))
	for i, p := range allProjects {
		projIdx[p.Name] = i
	}
	for _, s := range allSessions {
		if i, ok := projIdx[s.Project]; ok {
			p := &allProjects[i]
			p.MessageCount += s.MessageCount
			p.TotalTokens += s.TokenUsage.InputTokens + s.TokenUsage.OutputTokens +
				s.TokenUsage.CacheReadInputTokens + s.TokenUsage.CacheCreationInputTokens
			p.CostUSD += s.CostUSD
			if s.EndTime > p.LastActive {
				p.LastActive = s.EndTime
			}
		}
	}

	sort.Slice(allSessions, func(i, j int) bool {
		return allSessions[i].StartTime > allSessions[j].StartTime
	})
	sort.Slice(allProjects, func(i, j int) bool {
		return allProjects[i].CostUSD > allProjects[j].CostUSD
	})

	return allSessions, allProjects
}
