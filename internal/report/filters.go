package report

import (
	"fmt"
	"sort"
	"strings"

	"github.com/sarveshkapre/endpoint-perf-agent/internal/anomaly"
)

func ApplyFilters(result AnalysisResult, minSeverity string, top int) (AnalysisResult, error) {
	out := result
	if out.TotalAnomalies == 0 {
		out.TotalAnomalies = len(out.Anomalies)
	}

	if minSeverity == "" {
		minSeverity = "low"
	}
	minSeverity = strings.ToLower(strings.TrimSpace(minSeverity))
	minRank, ok := severityRank(minSeverity)
	if !ok {
		return AnalysisResult{}, fmt.Errorf("unknown severity: %s (expected low|medium|high|critical)", minSeverity)
	}
	if top < 0 {
		return AnalysisResult{}, fmt.Errorf("top must be greater than or equal to zero")
	}

	filtered := make([]anomaly.Anomaly, 0, len(out.Anomalies))
	for _, a := range out.Anomalies {
		rank, ok := severityRank(a.Severity)
		if !ok {
			continue
		}
		if rank >= minRank {
			filtered = append(filtered, a)
		}
	}

	if top > 0 && len(filtered) > top {
		sort.Slice(filtered, func(i, j int) bool { return abs(filtered[i].ZScore) > abs(filtered[j].ZScore) })
		filtered = filtered[:top]
	}

	out.Anomalies = filtered
	return out, nil
}

func severityRank(severity string) (int, bool) {
	switch severity {
	case "low":
		return 1, true
	case "medium":
		return 2, true
	case "high":
		return 3, true
	case "critical":
		return 4, true
	default:
		return 0, false
	}
}
