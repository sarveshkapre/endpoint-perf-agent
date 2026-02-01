package report

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/sarveshkapre/endpoint-perf-agent/internal/anomaly"
	"github.com/sarveshkapre/endpoint-perf-agent/internal/collector"
)

type AnalysisResult struct {
	Samples         int
	Duration        time.Duration
	WindowSize      int
	ZScoreThreshold float64
	TotalAnomalies  int
	Anomalies       []anomaly.Anomaly
	FirstTimestamp  time.Time
	LastTimestamp   time.Time
}

func Analyze(samples []collector.MetricSample, windowSize int, threshold float64) AnalysisResult {
	result := AnalysisResult{
		Samples:         len(samples),
		WindowSize:      windowSize,
		ZScoreThreshold: threshold,
	}
	if len(samples) == 0 {
		return result
	}

	ordered := samples
	if !isSortedByTimestamp(samples) {
		ordered = append([]collector.MetricSample(nil), samples...)
		sort.Slice(ordered, func(i, j int) bool { return ordered[i].Timestamp.Before(ordered[j].Timestamp) })
	}

	detector := anomaly.NewDetector(windowSize, threshold)
	result.FirstTimestamp = ordered[0].Timestamp
	result.LastTimestamp = ordered[len(ordered)-1].Timestamp
	result.Duration = result.LastTimestamp.Sub(result.FirstTimestamp)

	prev := ordered[0]
	for i := 1; i < len(ordered); i++ {
		current := ordered[i]
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

	result.TotalAnomalies = len(result.Anomalies)
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
	fmt.Fprintf(&b, "Window size: %d\n", result.WindowSize)
	fmt.Fprintf(&b, "Z-score threshold: %.2f\n", result.ZScoreThreshold)
	fmt.Fprintf(&b, "Anomalies: %d\n", len(result.Anomalies))
	if result.TotalAnomalies > 0 && result.TotalAnomalies != len(result.Anomalies) {
		fmt.Fprintf(&b, "Anomalies total: %d\n", result.TotalAnomalies)
	}
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
		fmt.Fprintf(&b, "- Duration: %s\n", result.Duration)
		fmt.Fprintf(&b, "- Window size: %d\n", result.WindowSize)
		fmt.Fprintf(&b, "- Z-score threshold: %.2f\n", result.ZScoreThreshold)
		fmt.Fprintf(&b, "- First sample: %s\n", result.FirstTimestamp.Format(time.RFC3339))
		fmt.Fprintf(&b, "- Last sample: %s\n", result.LastTimestamp.Format(time.RFC3339))
	}
	fmt.Fprintf(&b, "- Anomalies: %d\n", len(result.Anomalies))
	if result.TotalAnomalies > 0 && result.TotalAnomalies != len(result.Anomalies) {
		fmt.Fprintf(&b, "- Anomalies total: %d\n", result.TotalAnomalies)
	}
	b.WriteString("\n")

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

func FormatJSON(result AnalysisResult) ([]byte, error) {
	type analysisResultJSON struct {
		Samples         int               `json:"samples"`
		Duration        string            `json:"duration"`
		WindowSize      int               `json:"window_size"`
		ZScoreThreshold float64           `json:"zscore_threshold"`
		TotalAnomalies  int               `json:"anomalies_total"`
		FirstTimestamp  string            `json:"first_timestamp,omitempty"`
		LastTimestamp   string            `json:"last_timestamp,omitempty"`
		Anomalies       []anomaly.Anomaly `json:"anomalies"`
	}
	out := analysisResultJSON{
		Samples:         result.Samples,
		Duration:        result.Duration.String(),
		WindowSize:      result.WindowSize,
		ZScoreThreshold: result.ZScoreThreshold,
		TotalAnomalies:  result.TotalAnomalies,
		Anomalies:       result.Anomalies,
	}
	if !result.FirstTimestamp.IsZero() {
		out.FirstTimestamp = result.FirstTimestamp.Format(time.RFC3339)
	}
	if !result.LastTimestamp.IsZero() {
		out.LastTimestamp = result.LastTimestamp.Format(time.RFC3339)
	}
	return json.MarshalIndent(out, "", "  ")
}

func isSortedByTimestamp(samples []collector.MetricSample) bool {
	if len(samples) < 2 {
		return true
	}
	prev := samples[0].Timestamp
	for i := 1; i < len(samples); i++ {
		ts := samples[i].Timestamp
		if ts.Before(prev) {
			return false
		}
		prev = ts
	}
	return true
}

func abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
