package pricing

import "strings"

type Rate struct {
	Input      float64
	Output     float64
	CacheRead  float64
	CacheWrite float64
}

var table = map[string]Rate{
	"claude-sonnet-4-6":          {3.00, 15.00, 0.30, 3.75},
	"claude-opus-4-6":            {15.00, 75.00, 1.50, 18.75},
	"claude-opus-4-5-20251101":   {15.00, 75.00, 1.50, 18.75},
	"claude-sonnet-4-5-20250929": {3.00, 15.00, 0.30, 3.75},
	"claude-haiku-4-5":           {0.80, 4.00, 0.08, 1.00},
	"claude-haiku-4-5-20251001":  {0.80, 4.00, 0.08, 1.00},
}

var defaultRate = Rate{3.00, 15.00, 0.30, 3.75}

func ForModel(model string) Rate {
	if r, ok := table[model]; ok {
		return r
	}
	// Fuzzy prefix match (e.g. "claude-sonnet-4-6-..." → sonnet rate)
	for k, r := range table {
		base := strings.SplitN(k, "-2025", 2)[0]
		if strings.HasPrefix(model, base) {
			return r
		}
	}
	return defaultRate
}

func Compute(model string, input, output, cacheRead, cacheWrite int64) float64 {
	r := ForModel(model)
	return float64(input)/1e6*r.Input +
		float64(output)/1e6*r.Output +
		float64(cacheRead)/1e6*r.CacheRead +
		float64(cacheWrite)/1e6*r.CacheWrite
}
