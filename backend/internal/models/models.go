package models

// ─── Stats Cache ────────────────────────────────────────────────────────────

type DailyActivity struct {
	Date          string `json:"date"`
	MessageCount  int    `json:"messageCount"`
	SessionCount  int    `json:"sessionCount"`
	ToolCallCount int    `json:"toolCallCount"`
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
	SessionId    string `json:"sessionId"`
	Duration     int64  `json:"duration"`
	MessageCount int    `json:"messageCount"`
	Timestamp    string `json:"timestamp"`
}

type StatsCache struct {
	Version          int                        `json:"version"`
	LastComputedDate string                     `json:"lastComputedDate"`
	DailyActivity    []DailyActivity            `json:"dailyActivity"`
	DailyModelTokens []DailyModelTokens         `json:"dailyModelTokens"`
	ModelUsage       map[string]ModelUsageEntry `json:"modelUsage"`
	TotalSessions    int                        `json:"totalSessions"`
	TotalMessages    int                        `json:"totalMessages"`
	LongestSession   LongestSession             `json:"longestSession"`
	FirstSessionDate string                     `json:"firstSessionDate"`
	HourCounts       map[string]int             `json:"hourCounts"`
}

// ─── Session ─────────────────────────────────────────────────────────────────

type TokenUsage struct {
	InputTokens              int64 `json:"input_tokens"`
	OutputTokens             int64 `json:"output_tokens"`
	CacheReadInputTokens     int64 `json:"cache_read_input_tokens"`
	CacheCreationInputTokens int64 `json:"cache_creation_input_tokens"`
}

type SessionInfo struct {
	SessionId      string         `json:"sessionId"`
	Project        string         `json:"project"`
	Title          string         `json:"title"`
	LastPrompt     string         `json:"lastPrompt"`
	StartTime      string         `json:"startTime"`
	EndTime        string         `json:"endTime"`
	DurationMs     int64          `json:"durationMs"`
	MessageCount   int            `json:"messageCount"`
	ToolCalls      map[string]int `json:"toolCalls"`
	TotalToolCalls int            `json:"totalToolCalls"`
	CostUSD        float64        `json:"costUSD"`
	TokenUsage     TokenUsage     `json:"tokenUsage"`
	Model          string         `json:"model"`
	GitBranch      string         `json:"gitBranch"`
	EntryPoint     string         `json:"entryPoint"`
	Version        string         `json:"version"`
	HasThinking    bool           `json:"hasThinking"`
	ErrorCount     int            `json:"errorCount"`
	FileCount      int            `json:"fileCount"`
	ThinkingChars  int64          `json:"thinkingChars"`
	ApiTimeMs      int64          `json:"apiTimeMs"`
}

// ─── Project ─────────────────────────────────────────────────────────────────

type ProjectInfo struct {
	Name         string  `json:"name"`
	Path         string  `json:"path"`
	SessionCount int     `json:"sessionCount"`
	MessageCount int     `json:"messageCount"`
	TotalTokens  int64   `json:"totalTokens"`
	CostUSD      float64 `json:"costUSD"`
	LastActive   string  `json:"lastActive"`
	SizeBytes    int64   `json:"sizeBytes"`
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
