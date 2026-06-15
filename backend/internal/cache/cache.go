package cache

import (
	"log"
	"sync"
	"time"

	"claude-stats/internal/models"
	"claude-stats/internal/parsers"
)

// DataCache holds all parsed data and refreshes periodically.
type DataCache struct {
	mu          sync.RWMutex
	stats       *models.StatsCache
	sessions    []models.SessionInfo
	projects    []models.ProjectInfo
	derived     models.DerivedStats
	lastRefresh time.Time
	claudeDir   string
	ready       bool
}

func New(claudeDir string) *DataCache {
	return &DataCache{claudeDir: claudeDir}
}

func (c *DataCache) Refresh() {
	liveSessions, projects := parsers.ParseAllSessions(c.claudeDir)
	sessions := mergeAndPersist(c.claudeDir, liveSessions)
	stats := models.ComputeStatsFromSessions(sessions)
	derived := models.ComputeDerived(sessions)

	c.mu.Lock()
	c.stats = stats
	c.sessions = sessions
	c.projects = projects
	c.derived = derived
	c.lastRefresh = time.Now()
	c.ready = true
	c.mu.Unlock()

	log.Printf("cache: %d sessions, %d projects, thinking %.0f wpm",
		len(sessions), len(projects), derived.ThinkingWPM)
}

func (c *DataCache) StartAutoRefresh(interval time.Duration) {
	go func() {
		t := time.NewTicker(interval)
		for range t.C {
			c.Refresh()
		}
	}()
}

func (c *DataCache) Get() (*models.StatsCache, []models.SessionInfo, []models.ProjectInfo, models.DerivedStats) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.stats, c.sessions, c.projects, c.derived
}

func (c *DataCache) IsReady() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ready
}

func (c *DataCache) LastRefresh() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lastRefresh
}
