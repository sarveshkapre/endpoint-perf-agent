package report

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/sarveshkapre/endpoint-perf-agent/internal/anomaly"
	"github.com/sarveshkapre/endpoint-perf-agent/internal/collector"
)

type AnalysisResult struct {
	Samples        int
	Duration       time.Duration
	Anomalies      []anomaly.Anomaly
	FirstTimestamp time.Time
	LastTimestamp  time.Time
}

func Analyze(samples []collector.MetricSample, windowSize int, threshold float64) AnalysisResult {
	result := AnalysisResult{Samples: len(samples)}
	if len(samples) == 0 {
		return result
	}

	detector := anomaly.NewDetector(windowSize, threshold)
	result.FirstTimestamp = samples[0].Timestamp
	result.LastTimestamp = samples[len(samples)-1].Timestamp
	result.Duration = result.LastTimestamp.Sub(result.FirstTimestamp)

	prev := samples[0]
	for i := 1; i < len(samples); i++ {
		current := samples[i]
		dt := current.Timestamp.Sub(prev.Timestamp).Seconds()
		if dt <= 0 {
			dt = 1
		}

		metrics := map[string]float64{
			"cpu_percent":              current.CPUPercent,
			"mem_used_percent":         current.MemUsedPercent,
			"disk_used_percent":        current.DiskUsedPercent,
			"disk_read_bytes_per_sec":  float64(delta(current.DiskReadBytes, prev.DiskReadBytes)) / dt,
			"disk_write_bytes_per_sec": float64(delta(current.DiskWriteBytes, prev.DiskWriteBytes)) / dt,
			"net_rx_bytes_per_sec":     float64(delta(current.NetRxBytes, prev.NetRxBytes)) / dt,
			"net_tx_bytes_per_sec":     float64(delta(current.NetTxBytes, prev.NetTxBytes)) / dt,
		}

		for name, value := range metrics {
			if a := detector.Check(name, value); a != nil {
				result.Anomalies = append(result.Anomalies, *a)
			}
		}
		prev = current
	}

	return result
}

func delta(current, previous uint64) uint64 {
	if current < previous {
		return 0
	}
	return current - previous
}

func FormatSummary(result AnalysisResult) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Samples: %d\n", result.Samples)
	if result.Samples == 0 {
		return b.String()
	}
	fmt.Fprintf(&b, "Duration: %s\n", result.Duration)
	fmt.Fprintf(&b, "Anomalies: %d\n", len(result.Anomalies))
	if len(result.Anomalies) == 0 {
		return b.String()
	}
	sorted := make([]anomaly.Anomaly, len(result.Anomalies))
	copy(sorted, result.Anomalies)
	sort.Slice(sorted, func(i, j int) bool { return abs(sorted[i].ZScore) > abs(sorted[j].ZScore) })
	top := sorted
	if len(top) > 5 {
		top = top[:5]
	}
	b.WriteString("Top anomalies:\n")
	for _, a := range top {
		fmt.Fprintf(&b, "- %s: %.2f (z=%.2f, %s)\n", a.Name, a.Value, a.ZScore, a.Severity)
	}
	return b.String()
}

func FormatMarkdown(result AnalysisResult) string {
	var b strings.Builder
	b.WriteString("# Endpoint Performance Report\n\n")
	b.WriteString("## Summary\n")
	fmt.Fprintf(&b, "- Samples: %d\n", result.Samples)
	if result.Samples > 0 {
		fmt.Fprintf(&b, "- Window: %s\n", result.Duration)
		fmt.Fprintf(&b, "- First sample: %s\n", result.FirstTimestamp.Format(time.RFC3339))
		fmt.Fprintf(&b, "- Last sample: %s\n", result.LastTimestamp.Format(time.RFC3339))
	}
	fmt.Fprintf(&b, "- Anomalies: %d\n\n", len(result.Anomalies))

	if len(result.Anomalies) == 0 {
		b.WriteString("No anomalies detected.\n")
		return b.String()
	}

	b.WriteString("## Anomalies\n")
	sort.Slice(result.Anomalies, func(i, j int) bool { return abs(result.Anomalies[i].ZScore) > abs(result.Anomalies[j].ZScore) })
	for _, a := range result.Anomalies {
		fmt.Fprintf(&b, "- **%s**: value %.2f (baseline %.2f ± %.2f, z=%.2f, %s). %s\n", a.Name, a.Value, a.Mean, a.Stddev, a.ZScore, a.Severity, a.Explanation)
	}
	return b.String()
}

func abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
