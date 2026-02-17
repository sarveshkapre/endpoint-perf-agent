package anomaly

import (
	"fmt"
	"math"
	"sort"
	"time"
)

type ProcessAttribution struct {
	PID        int32   `json:"pid"`
	Name       string  `json:"name"`
	CPUPercent float64 `json:"cpu_percent"`
	RSSBytes   uint64  `json:"rss_bytes"`
}

const (
	RuleTypeZScore          = "zscore"
	RuleTypeStaticThreshold = "static_threshold"
	RuleTypePercentile      = "percentile_threshold"
)

type PercentileRule struct {
	Percentile float64
	Multiplier float64
}

type Anomaly struct {
	Name          string
	Timestamp     time.Time         `json:"timestamp,omitempty"`
	Labels        map[string]string `json:"labels,omitempty"`
	Value         float64
	RuleType      string  `json:"rule_type,omitempty"`
	Threshold     float64 `json:"threshold,omitempty"`
	Mean          float64
	Stddev        float64
	ZScore        float64
	Severity      string
	Explanation   string
	TopCPUProcess *ProcessAttribution `json:"top_cpu_process,omitempty"`
	TopMemProcess *ProcessAttribution `json:"top_mem_process,omitempty"`
}

type Detector struct {
	windowSize int
	threshold  float64
	history    map[string][]float64
}

func NewDetector(windowSize int, threshold float64) *Detector {
	if windowSize < 5 {
		windowSize = 5
	}
	if threshold <= 0 {
		threshold = 3.0
	}
	return &Detector{
		windowSize: windowSize,
		threshold:  threshold,
		history:    make(map[string][]float64),
	}
}

func (d *Detector) Check(name string, value float64) *Anomaly {
	history := d.history[name]
	mean, stddev := meanStddev(history)
	anomaly := (*Anomaly)(nil)
	if len(history) >= d.windowSize && stddev > 0 {
		z := (value - mean) / stddev
		if math.Abs(z) >= d.threshold {
			anomaly = &Anomaly{
				Name:        name,
				Value:       value,
				RuleType:    RuleTypeZScore,
				Mean:        mean,
				Stddev:      stddev,
				ZScore:      z,
				Severity:    severityFromZ(z),
				Explanation: explain(name, value, mean, z),
			}
		}
	}

	history = append(history, value)
	if len(history) > d.windowSize {
		history = history[len(history)-d.windowSize:]
	}
	d.history[name] = history

	return anomaly
}

func (d *Detector) History(name string) []float64 {
	history := d.history[name]
	if len(history) == 0 {
		return nil
	}
	out := make([]float64, len(history))
	copy(out, history)
	return out
}

func severityFromZ(z float64) string {
	abs := math.Abs(z)
	switch {
	case abs >= 6:
		return "critical"
	case abs >= 4:
		return "high"
	case abs >= 3:
		return "medium"
	default:
		return "low"
	}
}

func meanStddev(values []float64) (float64, float64) {
	if len(values) == 0 {
		return 0, 0
	}
	var sum float64
	for _, v := range values {
		sum += v
	}
	mean := sum / float64(len(values))
	var variance float64
	for _, v := range values {
		diff := v - mean
		variance += diff * diff
	}
	variance = variance / float64(len(values))
	return mean, math.Sqrt(variance)
}

func CheckStaticThreshold(name string, value float64, thresholds map[string]float64) *Anomaly {
	if len(thresholds) == 0 {
		return nil
	}
	threshold, ok := thresholds[name]
	if !ok || threshold <= 0 || value < threshold {
		return nil
	}
	exceedRatio := (value - threshold) / threshold
	return &Anomaly{
		Name:        name,
		Value:       value,
		RuleType:    RuleTypeStaticThreshold,
		Threshold:   threshold,
		Mean:        threshold,
		Stddev:      0,
		ZScore:      exceedRatio,
		Severity:    severityFromExceedRatio(exceedRatio),
		Explanation: explainStaticThreshold(name, value, threshold, exceedRatio),
	}
}

func CheckPercentileThreshold(name string, value float64, history []float64, rules map[string]PercentileRule) *Anomaly {
	if len(rules) == 0 {
		return nil
	}
	rule, ok := rules[name]
	if !ok {
		return nil
	}
	if err := ValidatePercentileRule(rule); err != nil {
		return nil
	}
	percentileValue, ok := percentile(history, rule.Percentile)
	if !ok || percentileValue <= 0 {
		return nil
	}
	threshold := percentileValue * rule.Multiplier
	if value < threshold {
		return nil
	}
	exceedRatio := (value - threshold) / threshold
	return &Anomaly{
		Name:        name,
		Value:       value,
		RuleType:    RuleTypePercentile,
		Threshold:   threshold,
		Mean:        percentileValue,
		Stddev:      0,
		ZScore:      exceedRatio,
		Severity:    severityFromExceedRatio(exceedRatio),
		Explanation: explainPercentileThreshold(name, value, rule.Percentile, percentileValue, rule.Multiplier, threshold, exceedRatio),
	}
}

func ValidatePercentileRule(rule PercentileRule) error {
	if math.IsNaN(rule.Percentile) || math.IsInf(rule.Percentile, 0) || rule.Percentile <= 0 || rule.Percentile > 100 {
		return fmt.Errorf("percentile must be in (0,100]")
	}
	if math.IsNaN(rule.Multiplier) || math.IsInf(rule.Multiplier, 0) || rule.Multiplier <= 0 {
		return fmt.Errorf("multiplier must be greater than zero")
	}
	return nil
}

func SelectHigherSeverity(a, b *Anomaly) *Anomaly {
	if a == nil {
		return b
	}
	if b == nil {
		return a
	}
	ar, _ := severityRank(a.Severity)
	br, _ := severityRank(b.Severity)
	switch {
	case br > ar:
		return b
	case br < ar:
		return a
	}

	az := math.Abs(a.ZScore)
	bz := math.Abs(b.ZScore)
	switch {
	case bz > az:
		return b
	case bz < az:
		return a
	}

	if a.RuleType == RuleTypeZScore {
		return a
	}
	if b.RuleType == RuleTypeZScore {
		return b
	}
	return a
}

func severityFromExceedRatio(ratio float64) string {
	switch {
	case ratio >= 1.0:
		return "critical"
	case ratio >= 0.5:
		return "high"
	case ratio >= 0.2:
		return "medium"
	default:
		return "low"
	}
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

func explainStaticThreshold(name string, value, threshold, exceedRatio float64) string {
	return fmt.Sprintf("Static threshold exceeded for %s: value %.2f is above %.2f (%.1f%% over threshold).", name, value, threshold, exceedRatio*100)
}

func explainPercentileThreshold(name string, value, percentileRank, percentileValue, multiplier, threshold, exceedRatio float64) string {
	return fmt.Sprintf("%s exceeded dynamic percentile threshold: value %.2f is above p%.0f baseline %.2f x %.2f (threshold %.2f, %.1f%% over threshold).", name, value, percentileRank, percentileValue, multiplier, threshold, exceedRatio*100)
}

func explain(name string, value, mean, z float64) string {
	sigma := math.Abs(z)
	trendUp := z >= 0

	switch name {
	case "cpu_percent":
		verb := "spiked"
		if !trendUp {
			verb = "dropped"
		}
		return fmt.Sprintf("CPU usage %s to %.1f%% (baseline %.1f%%, %.1fσ). Check for runaway processes, background jobs, or throttling.", verb, value, mean, sigma)
	case "mem_used_percent":
		verb := "rose"
		if !trendUp {
			verb = "fell"
		}
		return fmt.Sprintf("Memory usage %s to %.1f%% (baseline %.1f%%, %.1fσ). Look for leaks, large caches, or memory pressure.", verb, value, mean, sigma)
	case "disk_used_percent":
		verb := "rose"
		if !trendUp {
			verb = "fell"
		}
		return fmt.Sprintf("Disk usage %s to %.1f%% (baseline %.1f%%, %.1fσ). Investigate large writes, logs, or unexpected data growth.", verb, value, mean, sigma)
	case "disk_read_bytes_per_sec":
		verb := "jumped"
		if !trendUp {
			verb = "dropped"
		}
		return fmt.Sprintf("Disk read throughput %s to %.0f B/s (baseline %.0f B/s, %.1fσ). Possible causes: scans, backups, or stalled I/O.", verb, value, mean, sigma)
	case "disk_write_bytes_per_sec":
		verb := "jumped"
		if !trendUp {
			verb = "dropped"
		}
		return fmt.Sprintf("Disk write throughput %s to %.0f B/s (baseline %.0f B/s, %.1fσ). Check for log storms, sync jobs, or blocked writes.", verb, value, mean, sigma)
	case "net_rx_bytes_per_sec":
		verb := "spiked"
		if !trendUp {
			verb = "dropped"
		}
		return fmt.Sprintf("Inbound network %s to %.0f B/s (baseline %.0f B/s, %.1fσ). Verify unexpected downloads or large transfers.", verb, value, mean, sigma)
	case "net_tx_bytes_per_sec":
		verb := "spiked"
		if !trendUp {
			verb = "dropped"
		}
		return fmt.Sprintf("Outbound network %s to %.0f B/s (baseline %.0f B/s, %.1fσ). Look for uploads, backups, or exfil signals.", verb, value, mean, sigma)
	default:
		return fmt.Sprintf("Metric %s deviated from baseline (%.1fσ).", name, sigma)
	}
}

func percentile(values []float64, pct float64) (float64, bool) {
	if len(values) == 0 || pct <= 0 || pct > 100 {
		return 0, false
	}
	sorted := append([]float64(nil), values...)
	sort.Float64s(sorted)

	if len(sorted) == 1 {
		return sorted[0], true
	}

	rank := (pct / 100) * float64(len(sorted)-1)
	lower := int(math.Floor(rank))
	upper := int(math.Ceil(rank))
	if lower == upper {
		return sorted[lower], true
	}
	weight := rank - float64(lower)
	return sorted[lower] + (sorted[upper]-sorted[lower])*weight, true
}
