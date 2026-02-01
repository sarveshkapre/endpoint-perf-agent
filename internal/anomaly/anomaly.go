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
	switch name {
	case "cpu_percent":
		return fmt.Sprintf("CPU usage spiked to %.1f%% (baseline %.1f%%, %.1fσ). Check for runaway processes or batch workloads.", value, mean, z)
	case "mem_used_percent":
		return fmt.Sprintf("Memory usage hit %.1f%% (baseline %.1f%%, %.1fσ). Look for leaks, large caches, or memory pressure.", value, mean, z)
	case "disk_used_percent":
		return fmt.Sprintf("Disk usage rose to %.1f%% (baseline %.1f%%, %.1fσ). Investigate large writes, logs, or unexpected data growth.", value, mean, z)
	case "disk_read_bytes_per_sec":
		return fmt.Sprintf("Disk read throughput jumped to %.0f B/s (baseline %.0f B/s, %.1fσ). Possible causes: heavy scans or backups.", value, mean, z)
	case "disk_write_bytes_per_sec":
		return fmt.Sprintf("Disk write throughput jumped to %.0f B/s (baseline %.0f B/s, %.1fσ). Check for log storms or sync jobs.", value, mean, z)
	case "net_rx_bytes_per_sec":
		return fmt.Sprintf("Inbound network spiked to %.0f B/s (baseline %.0f B/s, %.1fσ). Verify unexpected downloads or large transfers.", value, mean, z)
	case "net_tx_bytes_per_sec":
		return fmt.Sprintf("Outbound network spiked to %.0f B/s (baseline %.0f B/s, %.1fσ). Look for uploads, backups, or exfil signals.", value, mean, z)
	default:
		return fmt.Sprintf("Metric %s deviated from baseline (%.1fσ).", name, z)
	}
}
