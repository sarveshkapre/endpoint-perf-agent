package anomaly

import (
	"fmt"
	"math"
)

type Anomaly struct {
	Name        string
	Value       float64
	Mean        float64
	Stddev      float64
	ZScore      float64
	Severity    string
	Explanation string
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
