package parsers

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"claude-stats/internal/models"
)

func ReadStatsCache(claudeDir string) *models.StatsCache {
	path := filepath.Join(claudeDir, "stats-cache.json")
	data, err := os.ReadFile(path)
	if err != nil {
		log.Printf("stats-cache.json: %v", err)
		return &models.StatsCache{}
	}
	var s models.StatsCache
	if err := json.Unmarshal(data, &s); err != nil {
		log.Printf("stats-cache.json parse: %v", err)
		return &models.StatsCache{}
	}
	return &s
}
