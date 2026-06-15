package pricing

import (
	"regexp"
	"strings"
)

type Rate struct {
	Input      float64 // $ per MTok
	Output     float64
	CacheRead  float64 // 0.1x input
	CacheWrite float64 // 5-min cache write, 1.25x input
}

// dateSuffix strips trailing date stamps like -20251001 or -20260615
var dateSuffix = regexp.MustCompile(`-20\d{6}.*$`)

func normalize(model string) string {
	return dateSuffix.ReplaceAllString(strings.ToLower(model), "")
}

// Prices from https://claude.com/pricing (June 2026)
// CacheWrite = 5-minute cache write (1.25x input); CacheRead = 0.1x input
var table = map[string]Rate{
	// Fable 5 / Mythos 5
	"claude-fable-5":  {10.00, 50.00, 1.00, 12.50},
	"claude-mythos-5": {10.00, 50.00, 1.00, 12.50},

	// Opus 4.8
	"claude-opus-4-8": {5.00, 25.00, 0.50, 6.25},

	// Opus 4.7
	"claude-opus-4-7": {5.00, 25.00, 0.50, 6.25},

	// Opus 4.6
	"claude-opus-4-6": {5.00, 25.00, 0.50, 6.25},

	// Opus 4.5
	"claude-opus-4-5": {5.00, 25.00, 0.50, 6.25},

	// Opus 4.1 (deprecated)
	"claude-opus-4-1": {15.00, 75.00, 1.50, 18.75},

	// Opus 4 (deprecated)
	"claude-opus-4": {15.00, 75.00, 1.50, 18.75},

	// Sonnet 4.6
	"claude-sonnet-4-6": {3.00, 15.00, 0.30, 3.75},

	// Sonnet 4.5
	"claude-sonnet-4-5": {3.00, 15.00, 0.30, 3.75},

	// Sonnet 4 (deprecated)
	"claude-sonnet-4": {3.00, 15.00, 0.30, 3.75},

	// Haiku 4.5
	"claude-haiku-4-5": {1.00, 5.00, 0.10, 1.25},

	// Haiku 3.5 (retired)
	"claude-haiku-3-5": {0.80, 4.00, 0.08, 1.00},
}

func ForModel(model string) Rate {
	key := normalize(model)
	if r, ok := table[key]; ok {
		return r
	}
	// Fallback: longest prefix match (handles unknown minor versions)
	best, bestLen := Rate{}, 0
	for k, r := range table {
		if strings.HasPrefix(key, k) && len(k) > bestLen {
			best, bestLen = r, len(k)
		}
	}
	if bestLen > 0 {
		return best
	}
	// Default to Sonnet 4.6 rates if truly unknown
	return table["claude-sonnet-4-6"]
}

func Compute(model string, input, output, cacheRead, cacheWrite int64) float64 {
	r := ForModel(model)
	return float64(input)/1e6*r.Input +
		float64(output)/1e6*r.Output +
		float64(cacheRead)/1e6*r.CacheRead +
		float64(cacheWrite)/1e6*r.CacheWrite
}
