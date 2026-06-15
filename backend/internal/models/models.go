package models

import (
	"sort"
	"strconv"
	"time"
)

// ─── Stats Cache ────────────────────────────────────────────────────────────

type DailyActivity struct {
	Date                 string `json:"date"`
	MessageCount         int    `json:"messageCount"`
	UserMessageCount     int    `json:"userMessageCount"`
	AssistantMessageCount int   `json:"assistantMessageCount"`
	SessionCount         int    `json:"sessionCount"`
	ToolCallCount        int    `json:"toolCallCount"`
}

type DailyModelTokens struct {
	Date          string           `json:"date"`
	TokensByModel map[string]int64 `json:"tokensByModel"`
}

type ModelUsageEntry struct {
	InputTokens              int64   `json:"inputTokens"`
	OutputTokens             int64   `json:"outputTokens"`
	CacheReadInputTokens     int64   `json:"cacheReadInputTokens"`
	CacheCreationInputTokens int64   `json:"cacheCreationInputTokens"`
	WebSearchRequests        int     `json:"webSearchRequests"`
	CostUSD                  float64 `json:"costUSD"`
}

type LongestSession struct {
	SessionId             string `json:"sessionId"`
	Duration              int64  `json:"duration"`
	MessageCount          int    `json:"messageCount"`
	UserMessageCount      int    `json:"userMessageCount"`
	AssistantMessageCount int    `json:"assistantMessageCount"`
	LinesAdded            int64  `json:"linesAdded"`
	Timestamp             string `json:"timestamp"`
}

type StatsCache struct {
	Version               int                        `json:"version"`
	LastComputedDate      string                     `json:"lastComputedDate"`
	DailyActivity         []DailyActivity            `json:"dailyActivity"`
	DailyModelTokens      []DailyModelTokens         `json:"dailyModelTokens"`
	ModelUsage            map[string]ModelUsageEntry `json:"modelUsage"`
	TotalSessions          int                        `json:"totalSessions"`
	TotalMessages          int                        `json:"totalMessages"`
	TotalUserMessages      int                        `json:"totalUserMessages"`
	TotalAssistantMessages int                        `json:"totalAssistantMessages"`
	TotalFilesEdited       int                        `json:"totalFilesEdited"`
	TotalLinesAdded        int64                      `json:"totalLinesAdded"`
	LongestSession        LongestSession             `json:"longestSession"`
	FirstSessionDate      string                     `json:"firstSessionDate"`
	HourCounts            map[string]int             `json:"hourCounts"`
}

// ─── Session ─────────────────────────────────────────────────────────────────

type TokenUsage struct {
	InputTokens              int64 `json:"input_tokens"`
	OutputTokens             int64 `json:"output_tokens"`
	CacheReadInputTokens     int64 `json:"cache_read_input_tokens"`
	CacheCreationInputTokens int64 `json:"cache_creation_input_tokens"`
}

type SessionInfo struct {
	SessionId             string         `json:"sessionId"`
	Project               string         `json:"project"`
	ProjectPath           string         `json:"projectPath,omitempty"`
	Title                 string         `json:"title"`
	LastPrompt            string         `json:"lastPrompt"`
	StartTime             string         `json:"startTime"`
	EndTime               string         `json:"endTime"`
	DurationMs            int64          `json:"durationMs"`
	MessageCount          int            `json:"messageCount"`
	UserMessageCount      int            `json:"userMessageCount"`
	AssistantMessageCount int            `json:"assistantMessageCount"`
	ToolCalls             map[string]int `json:"toolCalls"`
	TotalToolCalls        int            `json:"totalToolCalls"`
	CostUSD        float64        `json:"costUSD"`
	TokenUsage     TokenUsage     `json:"tokenUsage"`
	Model          string         `json:"model"`
	GitBranch      string         `json:"gitBranch"`
	EntryPoint     string         `json:"entryPoint"`
	Version        string         `json:"version"`
	HasThinking    bool           `json:"hasThinking"`
	ErrorCount     int            `json:"errorCount"`
	FileCount      int            `json:"fileCount"`
	ThinkingChars       int64          `json:"thinkingChars"`
	ApiTimeMs           int64          `json:"apiTimeMs"`
	LinesAdded          int64          `json:"linesAdded"`
	WebSearchCount      int            `json:"webSearchCount"`
	WebFetchCount       int            `json:"webFetchCount"`
	WebSearchQueries    []string       `json:"webSearchQueries,omitempty"`
	WebFetchDomains     []string       `json:"webFetchDomains,omitempty"`
	BashNetworkCounts   map[string]int `json:"bashNetworkCounts,omitempty"`
}

// ─── Project ─────────────────────────────────────────────────────────────────

type ProjectInfo struct {
	Name                  string  `json:"name"`
	Path                  string  `json:"path"`
	SessionCount          int     `json:"sessionCount"`
	MessageCount          int     `json:"messageCount"`
	UserMessageCount      int     `json:"userMessageCount"`
	AssistantMessageCount int     `json:"assistantMessageCount"`
	TotalTokens           int64   `json:"totalTokens"`
	CostUSD               float64 `json:"costUSD"`
	LastActive            string  `json:"lastActive"`
	SizeBytes             int64   `json:"sizeBytes"`
}

// ─── Storage ──────────────────────────────────────────────────────────────────

type StorageInfo struct {
	TotalBytes int64            `json:"totalBytes"`
	ByDir      map[string]int64 `json:"byDir"`
	ByProject  []ProjectStorage `json:"byProject"`
}

type ProjectStorage struct {
	Project  string `json:"project"`
	Bytes    int64  `json:"bytes"`
	Sessions int    `json:"sessions"`
}

// ─── API Responses ────────────────────────────────────────────────────────────

type OverviewResponse struct {
	Stats            *StatsCache          `json:"stats"`
	ComputedCosts    map[string]float64   `json:"computedCosts"`
	TotalCostUSD     float64              `json:"totalCostUSD"`
	ThinkingWPM      float64              `json:"thinkingWPM"`
}

// ─── Global Stats (derived from sessions) ─────────────────────────────────────

type DerivedStats struct {
	TotalToolCalls  int64
	TotalErrors     int64
	ThinkingChars   int64
	ApiTimeMs       int64
	ThinkingWPM     float64
}

// ComputeStatsFromSessions rebuilds StatsCache entirely from live session data,
// bypassing the stale stats-cache.json written by Claude Code.
func ComputeStatsFromSessions(sessions []SessionInfo) *StatsCache {
	if len(sessions) == 0 {
		return &StatsCache{}
	}

	modelUsage := make(map[string]ModelUsageEntry)
	dailyActivityMap := make(map[string]*DailyActivity)
	dailyModelTokensMap := make(map[string]map[string]int64)
	hourCounts := make(map[string]int)

	var totalMessages, totalUserMessages, totalAssistantMessages, totalFilesEdited int
	var totalLinesAdded int64
	var firstDate, lastDate string
	var longestSession LongestSession

	for _, s := range sessions {
		if s.StartTime == "" {
			continue
		}
		t, err := time.Parse(time.RFC3339, s.StartTime)
		if err != nil {
			continue
		}
		date := t.UTC().Format("2006-01-02")
		hour := strconv.Itoa(t.UTC().Hour())

		hourCounts[hour]++
		totalMessages += s.MessageCount
		totalUserMessages += s.UserMessageCount
		totalAssistantMessages += s.AssistantMessageCount
		totalFilesEdited += s.FileCount
		totalLinesAdded += s.LinesAdded

		if firstDate == "" || date < firstDate {
			firstDate = date
		}
		if date > lastDate {
			lastDate = date
		}

		if s.MessageCount > longestSession.MessageCount {
			longestSession = LongestSession{
				SessionId:             s.SessionId,
				MessageCount:          s.MessageCount,
				UserMessageCount:      s.UserMessageCount,
				AssistantMessageCount: s.AssistantMessageCount,
				LinesAdded:            s.LinesAdded,
				Duration:              s.DurationMs,
				Timestamp:             s.StartTime,
			}
		}

		// Daily activity
		if _, ok := dailyActivityMap[date]; !ok {
			dailyActivityMap[date] = &DailyActivity{Date: date}
		}
		dailyActivityMap[date].MessageCount += s.MessageCount
		dailyActivityMap[date].UserMessageCount += s.UserMessageCount
		dailyActivityMap[date].AssistantMessageCount += s.AssistantMessageCount
		dailyActivityMap[date].SessionCount++
		dailyActivityMap[date].ToolCallCount += s.TotalToolCalls

		// Model usage
		model := s.Model
		if model == "" {
			model = "unknown"
		}
		u := modelUsage[model]
		u.InputTokens += s.TokenUsage.InputTokens
		u.OutputTokens += s.TokenUsage.OutputTokens
		u.CacheReadInputTokens += s.TokenUsage.CacheReadInputTokens
		u.CacheCreationInputTokens += s.TokenUsage.CacheCreationInputTokens
		u.CostUSD += s.CostUSD
		modelUsage[model] = u

		// Daily model tokens
		if _, ok := dailyModelTokensMap[date]; !ok {
			dailyModelTokensMap[date] = make(map[string]int64)
		}
		dailyModelTokensMap[date][model] += s.TokenUsage.InputTokens +
			s.TokenUsage.OutputTokens + s.TokenUsage.CacheReadInputTokens
	}

	// Sort dates
	dates := make([]string, 0, len(dailyActivityMap))
	for d := range dailyActivityMap {
		dates = append(dates, d)
	}
	sort.Strings(dates)

	dailyActivity := make([]DailyActivity, 0, len(dates))
	dailyModelTokens := make([]DailyModelTokens, 0, len(dates))
	for _, d := range dates {
		dailyActivity = append(dailyActivity, *dailyActivityMap[d])
		dailyModelTokens = append(dailyModelTokens, DailyModelTokens{
			Date:          d,
			TokensByModel: dailyModelTokensMap[d],
		})
	}

	return &StatsCache{
		LastComputedDate:       lastDate,
		DailyActivity:          dailyActivity,
		DailyModelTokens:       dailyModelTokens,
		ModelUsage:             modelUsage,
		TotalSessions:          len(sessions),
		TotalMessages:          totalMessages,
		TotalUserMessages:      totalUserMessages,
		TotalAssistantMessages: totalAssistantMessages,
		TotalFilesEdited:       totalFilesEdited,
		TotalLinesAdded:        totalLinesAdded,
		LongestSession:         longestSession,
		FirstSessionDate:       firstDate,
		HourCounts:             hourCounts,
	}
}

func ComputeDerived(sessions []SessionInfo) DerivedStats {
	var d DerivedStats
	for _, s := range sessions {
		d.TotalToolCalls += int64(s.TotalToolCalls)
		d.TotalErrors += int64(s.ErrorCount)
		d.ThinkingChars += s.ThinkingChars
		d.ApiTimeMs += s.ApiTimeMs
	}
	if d.ApiTimeMs > 0 {
		minutes := float64(d.ApiTimeMs) / 60000.0
		words := float64(d.ThinkingChars) / 5.0
		if minutes > 0 {
			d.ThinkingWPM = words / minutes
		}
	}
	return d
}
