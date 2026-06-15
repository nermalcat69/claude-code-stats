package cache

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"sort"

	"claude-stats/internal/models"
)

const persistFileName = "claude-code-stats.json"

// SessionRecord is a lightweight snapshot of one session stored in our cache.
type SessionRecord struct {
	SessionId             string  `json:"sessionId"`
	Project               string  `json:"project"`
	ProjectPath           string  `json:"projectPath,omitempty"`
	StartTime             string  `json:"startTime"`
	EndTime               string  `json:"endTime"`
	DurationMs            int64   `json:"durationMs"`
	MessageCount          int     `json:"messageCount"`
	UserMessageCount      int     `json:"userMessageCount"`
	AssistantMessageCount int     `json:"assistantMessageCount"`
	ToolCalls             map[string]int `json:"toolCalls,omitempty"`
	TotalToolCalls        int            `json:"totalToolCalls"`
	LinesAdded            int64   `json:"linesAdded"`
	FileCount             int     `json:"fileCount"`
	Model                 string  `json:"model"`
	InputTokens           int64   `json:"inputTokens"`
	OutputTokens          int64   `json:"outputTokens"`
	CacheReadInputTokens  int64   `json:"cacheReadInputTokens"`
	CacheCreationTokens   int64   `json:"cacheCreationTokens"`
	CostUSD               float64 `json:"costUSD"`
	ThinkingChars       int64          `json:"thinkingChars"`
	ApiTimeMs           int64          `json:"apiTimeMs"`
	WebSearchCount      int            `json:"webSearchCount"`
	WebFetchCount       int            `json:"webFetchCount"`
	WebSearchQueries    []string       `json:"webSearchQueries,omitempty"`
	WebFetchDomains     []string       `json:"webFetchDomains,omitempty"`
	BashNetworkCounts   map[string]int `json:"bashNetworkCounts,omitempty"`
}

type PersistentCache struct {
	Version  int                      `json:"version"`
	Sessions map[string]SessionRecord `json:"sessions"`
}

func persistPath(claudeDir string) string {
	return filepath.Join(claudeDir, persistFileName)
}

func loadPersistent(claudeDir string) PersistentCache {
	pc := PersistentCache{Version: 1, Sessions: make(map[string]SessionRecord)}
	data, err := os.ReadFile(persistPath(claudeDir))
	if err != nil {
		return pc // first run
	}
	if err := json.Unmarshal(data, &pc); err != nil {
		log.Printf("persistent cache parse error: %v — starting fresh", err)
		pc.Sessions = make(map[string]SessionRecord)
	}
	return pc
}

func savePersistent(claudeDir string, pc PersistentCache) {
	data, err := json.Marshal(pc)
	if err != nil {
		log.Printf("persistent cache marshal: %v", err)
		return
	}
	if err := os.WriteFile(persistPath(claudeDir), data, 0644); err != nil {
		log.Printf("persistent cache write: %v", err)
	}
}

func sessionToRecord(s models.SessionInfo) SessionRecord {
	return SessionRecord{
		SessionId:             s.SessionId,
		Project:               s.Project,
		ProjectPath:           s.ProjectPath,
		StartTime:             s.StartTime,
		EndTime:               s.EndTime,
		DurationMs:            s.DurationMs,
		MessageCount:          s.MessageCount,
		UserMessageCount:      s.UserMessageCount,
		AssistantMessageCount: s.AssistantMessageCount,
		ToolCalls:             s.ToolCalls,
		TotalToolCalls:        s.TotalToolCalls,
		LinesAdded:            s.LinesAdded,
		FileCount:             s.FileCount,
		Model:                 s.Model,
		InputTokens:           s.TokenUsage.InputTokens,
		OutputTokens:          s.TokenUsage.OutputTokens,
		CacheReadInputTokens:  s.TokenUsage.CacheReadInputTokens,
		CacheCreationTokens:   s.TokenUsage.CacheCreationInputTokens,
		CostUSD:               s.CostUSD,
		ThinkingChars:         s.ThinkingChars,
		ApiTimeMs:             s.ApiTimeMs,
		WebSearchCount:        s.WebSearchCount,
		WebFetchCount:         s.WebFetchCount,
		WebSearchQueries:      s.WebSearchQueries,
		WebFetchDomains:       s.WebFetchDomains,
		BashNetworkCounts:     s.BashNetworkCounts,
	}
}

func recordToSession(r SessionRecord) models.SessionInfo {
	return models.SessionInfo{
		SessionId:             r.SessionId,
		Project:               r.Project,
		ProjectPath:           r.ProjectPath,
		StartTime:             r.StartTime,
		EndTime:               r.EndTime,
		DurationMs:            r.DurationMs,
		MessageCount:          r.MessageCount,
		UserMessageCount:      r.UserMessageCount,
		AssistantMessageCount: r.AssistantMessageCount,
		ToolCalls:             r.ToolCalls,
		TotalToolCalls:        r.TotalToolCalls,
		LinesAdded:            r.LinesAdded,
		FileCount:             r.FileCount,
		Model:                 r.Model,
		TokenUsage: models.TokenUsage{
			InputTokens:              r.InputTokens,
			OutputTokens:             r.OutputTokens,
			CacheReadInputTokens:     r.CacheReadInputTokens,
			CacheCreationInputTokens: r.CacheCreationTokens,
		},
		CostUSD:           r.CostUSD,
		ThinkingChars:     r.ThinkingChars,
		ApiTimeMs:         r.ApiTimeMs,
		WebSearchCount:    r.WebSearchCount,
		WebFetchCount:     r.WebFetchCount,
		WebSearchQueries:  r.WebSearchQueries,
		WebFetchDomains:   r.WebFetchDomains,
		BashNetworkCounts: r.BashNetworkCounts,
	}
}

// mergeAndPersist upserts live sessions into the persistent store and returns
// the complete session list (live + any historical records not in live).
func mergeAndPersist(claudeDir string, liveSessions []models.SessionInfo) []models.SessionInfo {
	pc := loadPersistent(claudeDir)

	// Upsert every live session (live data is always more accurate)
	liveIDs := make(map[string]bool, len(liveSessions))
	for _, s := range liveSessions {
		if s.SessionId == "" {
			continue
		}
		liveIDs[s.SessionId] = true
		pc.Sessions[s.SessionId] = sessionToRecord(s)
	}

	// Convert full index back to []SessionInfo
	all := make([]models.SessionInfo, 0, len(pc.Sessions))
	for _, r := range pc.Sessions {
		all = append(all, recordToSession(r))
	}

	sort.Slice(all, func(i, j int) bool {
		return all[i].StartTime > all[j].StartTime
	})

	savePersistent(claudeDir, pc)

	historical := len(pc.Sessions) - len(liveIDs)
	if historical > 0 {
		log.Printf("persistent cache: %d live + %d historical sessions restored", len(liveIDs), historical)
	}

	return all
}
