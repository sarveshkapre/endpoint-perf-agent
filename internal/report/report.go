package report

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/sarveshkapre/endpoint-perf-agent/internal/anomaly"
	"github.com/sarveshkapre/endpoint-perf-agent/internal/collector"
)

const (
	minWindowSize    = 5
	defaultThreshold = 3.0
)

type MetricStats struct {
	Count  int     `json:"count"`
	Mean   float64 `json:"mean"`
	Stddev float64 `json:"stddev"`
	Min    float64 `json:"min"`
	Max    float64 `json:"max"`
}

type AnalysisResult struct {
	Samples         int
	Duration        time.Duration
	WindowSize      int
	ZScoreThreshold float64
	HostID          string
	TotalAnomalies  int
	Anomalies       []anomaly.Anomaly
	Baselines       map[string]MetricStats
	FirstTimestamp  time.Time
	LastTimestamp   time.Time
}

func Analyze(samples []collector.MetricSample, windowSize int, threshold float64) AnalysisResult {
	windowSize, threshold = NormalizeParams(windowSize, threshold)
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

	result.HostID = stableHostID(ordered)

	detector := anomaly.NewDetector(windowSize, threshold)
	result.FirstTimestamp = ordered[0].Timestamp
	result.LastTimestamp = ordered[len(ordered)-1].Timestamp
	result.Duration = result.LastTimestamp.Sub(result.FirstTimestamp)

	metricValues := map[string][]float64{}
	firstFamilies := collector.DefaultMetricFamilies()
	if ordered[0].MetricFamilies != nil {
		firstFamilies = *ordered[0].MetricFamilies
	}
	if firstFamilies.CPU {
		metricValues["cpu_percent"] = []float64{ordered[0].CPUPercent}
	}
	if firstFamilies.Mem {
		metricValues["mem_used_percent"] = []float64{ordered[0].MemUsedPercent}
	}
	if firstFamilies.Disk {
		metricValues["disk_used_percent"] = []float64{ordered[0].DiskUsedPercent}
	}

	prev := ordered[0]
	for i := 1; i < len(ordered); i++ {
		current := ordered[i]
		currentFamilies := collector.DefaultMetricFamilies()
		if current.MetricFamilies != nil {
			currentFamilies = *current.MetricFamilies
		}
		prevFamilies := collector.DefaultMetricFamilies()
		if prev.MetricFamilies != nil {
			prevFamilies = *prev.MetricFamilies
		}
		dt := current.Timestamp.Sub(prev.Timestamp).Seconds()
		if dt <= 0 {
			dt = 1
		}

		metrics := map[string]float64{}
		if currentFamilies.CPU {
			metrics["cpu_percent"] = current.CPUPercent
		}
		if currentFamilies.Mem {
			metrics["mem_used_percent"] = current.MemUsedPercent
		}
		if currentFamilies.Disk {
			metrics["disk_used_percent"] = current.DiskUsedPercent
			if prevFamilies.Disk {
				metrics["disk_read_bytes_per_sec"] = float64(delta(current.DiskReadBytes, prev.DiskReadBytes)) / dt
				metrics["disk_write_bytes_per_sec"] = float64(delta(current.DiskWriteBytes, prev.DiskWriteBytes)) / dt
			}
		}
		if currentFamilies.Net && prevFamilies.Net {
			metrics["net_rx_bytes_per_sec"] = float64(delta(current.NetRxBytes, prev.NetRxBytes)) / dt
			metrics["net_tx_bytes_per_sec"] = float64(delta(current.NetTxBytes, prev.NetTxBytes)) / dt
		}

		for name, value := range metrics {
			metricValues[name] = append(metricValues[name], value)
			if a := detector.Check(name, value); a != nil {
				a.Timestamp = current.Timestamp
				a.TopCPUProcess = toAnomalyProcess(current.TopCPUProcess)
				a.TopMemProcess = toAnomalyProcess(current.TopMemProcess)
				result.Anomalies = append(result.Anomalies, *a)
			}
		}
		prev = current
	}

	result.Baselines = make(map[string]MetricStats, len(metricValues))
	for name, values := range metricValues {
		if len(values) == 0 {
			continue
		}
		result.Baselines[name] = computeStats(values)
	}

	result.TotalAnomalies = len(result.Anomalies)
	return result
}

func NormalizeParams(windowSize int, threshold float64) (int, float64) {
	if windowSize < minWindowSize {
		windowSize = minWindowSize
	}
	if threshold <= 0 {
		threshold = defaultThreshold
	}
	return windowSize, threshold
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
	if result.HostID != "" {
		fmt.Fprintf(&b, "Host: %s\n", result.HostID)
	}
	fmt.Fprintf(&b, "Duration: %s\n", result.Duration)
	fmt.Fprintf(&b, "Window size: %d\n", result.WindowSize)
	fmt.Fprintf(&b, "Z-score threshold: %.2f\n", result.ZScoreThreshold)
	if len(result.Baselines) > 0 {
		fmt.Fprintf(&b, "Baselines: %d metrics\n", len(result.Baselines))
	}
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
		fmt.Fprintf(&b, "- %s: %.2f (z=%.2f, %s)%s\n", a.Name, a.Value, a.ZScore, a.Severity, formatAnomalyContextInline(a))
	}
	return b.String()
}

func FormatMarkdown(result AnalysisResult) string {
	var b strings.Builder
	b.WriteString("# Endpoint Performance Report\n\n")
	b.WriteString("## Summary\n")
	fmt.Fprintf(&b, "- Samples: %d\n", result.Samples)
	if result.Samples > 0 {
		if result.HostID != "" {
			fmt.Fprintf(&b, "- Host: %s\n", result.HostID)
		}
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

	if len(result.Baselines) > 0 {
		b.WriteString("## Baselines\n")
		b.WriteString("| Metric | Mean | Stddev | Min | Max | Count |\n")
		b.WriteString("| --- | ---: | ---: | ---: | ---: | ---: |\n")
		for _, name := range orderedMetricNames(result.Baselines) {
			stats := result.Baselines[name]
			fmt.Fprintf(&b, "| %s | %s | %s | %s | %s | %d |\n",
				name,
				formatMetricValue(name, stats.Mean),
				formatMetricValue(name, stats.Stddev),
				formatMetricValue(name, stats.Min),
				formatMetricValue(name, stats.Max),
				stats.Count,
			)
		}
		b.WriteString("\n")
	}

	if len(result.Anomalies) == 0 {
		b.WriteString("No anomalies detected.\n")
		return b.String()
	}

	b.WriteString("## Anomalies\n")
	sort.Slice(result.Anomalies, func(i, j int) bool { return abs(result.Anomalies[i].ZScore) > abs(result.Anomalies[j].ZScore) })
	for _, a := range result.Anomalies {
		fmt.Fprintf(&b, "- **%s**: value %s (baseline %s ± %s, z=%.2f, %s). %s%s\n",
			a.Name,
			formatMetricValue(a.Name, a.Value),
			formatMetricValue(a.Name, a.Mean),
			formatMetricValue(a.Name, a.Stddev),
			a.ZScore,
			a.Severity,
			a.Explanation,
			formatAnomalyContextParagraph(a),
		)
	}
	return b.String()
}

func FormatJSON(result AnalysisResult) ([]byte, error) {
	type analysisResultJSON struct {
		Samples         int                    `json:"samples"`
		Duration        string                 `json:"duration"`
		WindowSize      int                    `json:"window_size"`
		ZScoreThreshold float64                `json:"zscore_threshold"`
		HostID          string                 `json:"host_id,omitempty"`
		TotalAnomalies  int                    `json:"anomalies_total"`
		FirstTimestamp  string                 `json:"first_timestamp,omitempty"`
		LastTimestamp   string                 `json:"last_timestamp,omitempty"`
		Anomalies       []anomaly.Anomaly      `json:"anomalies"`
		Baselines       map[string]MetricStats `json:"baselines,omitempty"`
	}
	out := analysisResultJSON{
		Samples:         result.Samples,
		Duration:        result.Duration.String(),
		WindowSize:      result.WindowSize,
		ZScoreThreshold: result.ZScoreThreshold,
		HostID:          result.HostID,
		TotalAnomalies:  result.TotalAnomalies,
		Anomalies:       result.Anomalies,
		Baselines:       result.Baselines,
	}
	if !result.FirstTimestamp.IsZero() {
		out.FirstTimestamp = result.FirstTimestamp.Format(time.RFC3339)
	}
	if !result.LastTimestamp.IsZero() {
		out.LastTimestamp = result.LastTimestamp.Format(time.RFC3339)
	}
	return json.MarshalIndent(out, "", "  ")
}

func stableHostID(samples []collector.MetricSample) string {
	host := ""
	for _, s := range samples {
		if s.HostID == "" {
			continue
		}
		if host == "" {
			host = s.HostID
			continue
		}
		if s.HostID != host {
			return ""
		}
	}
	return host
}

func toAnomalyProcess(p *collector.ProcessAttribution) *anomaly.ProcessAttribution {
	if p == nil {
		return nil
	}
	return &anomaly.ProcessAttribution{
		PID:        p.PID,
		Name:       p.Name,
		CPUPercent: p.CPUPercent,
		RSSBytes:   p.RSSBytes,
	}
}

func computeStats(values []float64) MetricStats {
	if len(values) == 0 {
		return MetricStats{}
	}
	stats := MetricStats{
		Count: len(values),
		Min:   values[0],
		Max:   values[0],
	}
	var sum float64
	for _, v := range values {
		sum += v
		if v < stats.Min {
			stats.Min = v
		}
		if v > stats.Max {
			stats.Max = v
		}
	}
	stats.Mean = sum / float64(len(values))

	var variance float64
	for _, v := range values {
		diff := v - stats.Mean
		variance += diff * diff
	}
	variance = variance / float64(len(values))
	stats.Stddev = math.Sqrt(variance)
	return stats
}

func orderedMetricNames(m map[string]MetricStats) []string {
	preferred := []string{
		"cpu_percent",
		"mem_used_percent",
		"disk_used_percent",
		"disk_read_bytes_per_sec",
		"disk_write_bytes_per_sec",
		"net_rx_bytes_per_sec",
		"net_tx_bytes_per_sec",
	}

	seen := make(map[string]bool, len(m))
	out := make([]string, 0, len(m))
	for _, name := range preferred {
		if _, ok := m[name]; ok {
			out = append(out, name)
			seen[name] = true
		}
	}
	rest := make([]string, 0, len(m))
	for name := range m {
		if !seen[name] {
			rest = append(rest, name)
		}
	}
	sort.Strings(rest)
	return append(out, rest...)
}

func formatMetricValue(name string, v float64) string {
	switch {
	case strings.HasSuffix(name, "_percent"):
		return fmt.Sprintf("%.1f%%", v)
	case strings.HasSuffix(name, "_bytes_per_sec"):
		return fmt.Sprintf("%s/s", humanBytes(v))
	default:
		return fmt.Sprintf("%.2f", v)
	}
}

func formatAnomalyContextInline(a anomaly.Anomaly) string {
	parts := make([]string, 0, 3)
	if !a.Timestamp.IsZero() {
		parts = append(parts, fmt.Sprintf("at %s", a.Timestamp.Format(time.RFC3339)))
	}
	if a.TopCPUProcess != nil {
		parts = append(parts, fmt.Sprintf("top CPU %s", formatProcessInline(*a.TopCPUProcess)))
	}
	if a.TopMemProcess != nil {
		parts = append(parts, fmt.Sprintf("top MEM %s", formatProcessInline(*a.TopMemProcess)))
	}
	if len(parts) == 0 {
		return ""
	}
	return " [" + strings.Join(parts, "; ") + "]"
}

func formatAnomalyContextParagraph(a anomaly.Anomaly) string {
	parts := make([]string, 0, 3)
	if !a.Timestamp.IsZero() {
		parts = append(parts, fmt.Sprintf("Observed at %s.", a.Timestamp.Format(time.RFC3339)))
	}
	if a.TopCPUProcess != nil {
		parts = append(parts, fmt.Sprintf("Top CPU process: %s.", formatProcessDetailed(*a.TopCPUProcess)))
	}
	if a.TopMemProcess != nil {
		parts = append(parts, fmt.Sprintf("Top memory process: %s.", formatProcessDetailed(*a.TopMemProcess)))
	}
	if len(parts) == 0 {
		return ""
	}
	return " " + strings.Join(parts, " ")
}

func formatProcessInline(p anomaly.ProcessAttribution) string {
	return fmt.Sprintf("%s(pid=%d)", p.Name, p.PID)
}

func formatProcessDetailed(p anomaly.ProcessAttribution) string {
	return fmt.Sprintf("%s (pid %d, cpu %.1f%%, rss %s)", p.Name, p.PID, p.CPUPercent, humanBytes(float64(p.RSSBytes)))
}

func humanBytes(v float64) string {
	if v < 0 {
		v = -v
	}
	units := []string{"B", "KiB", "MiB", "GiB", "TiB"}
	unit := 0
	for v >= 1024 && unit < len(units)-1 {
		v /= 1024
		unit++
	}
	if unit == 0 {
		return fmt.Sprintf("%.0f %s", v, units[unit])
	}
	if v >= 10 {
		return fmt.Sprintf("%.1f %s", v, units[unit])
	}
	return fmt.Sprintf("%.2f %s", v, units[unit])
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
