package report

import (
	"strings"

	"github.com/sarveshkapre/endpoint-perf-agent/internal/anomaly"
	"github.com/sarveshkapre/endpoint-perf-agent/internal/collector"
)

// FilterByMetricFamilies filters baselines and anomalies to only the enabled
// metric families. This is an output-only filter (it does not change how
// analysis is computed).
func FilterByMetricFamilies(result AnalysisResult, families collector.MetricFamilies) AnalysisResult {
	out := result

	if len(out.Baselines) > 0 {
		filtered := make(map[string]MetricStats, len(out.Baselines))
		for name, stats := range out.Baselines {
			if metricEnabledByFamily(name, families) {
				filtered[name] = stats
			}
		}
		out.Baselines = filtered
	}

	if len(out.Anomalies) > 0 {
		filtered := make([]anomaly.Anomaly, 0, len(out.Anomalies))
		for _, a := range out.Anomalies {
			if metricEnabledByFamily(a.Name, families) {
				filtered = append(filtered, a)
			}
		}
		out.Anomalies = filtered
	}

	return out
}

func metricEnabledByFamily(name string, families collector.MetricFamilies) bool {
	name = strings.ToLower(strings.TrimSpace(name))
	switch {
	case strings.HasPrefix(name, "cpu_"):
		return families.CPU
	case strings.HasPrefix(name, "mem_"):
		return families.Mem
	case strings.HasPrefix(name, "disk_"):
		return families.Disk
	case strings.HasPrefix(name, "net_"):
		return families.Net
	default:
		// Unknown metric name: keep it so we don't hide future metrics by default.
		return true
	}
}
