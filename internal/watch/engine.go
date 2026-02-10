package watch

import (
	"fmt"
	"strings"
	"time"

	"github.com/sarveshkapre/endpoint-perf-agent/internal/alert"
	"github.com/sarveshkapre/endpoint-perf-agent/internal/anomaly"
	"github.com/sarveshkapre/endpoint-perf-agent/internal/collector"
	"github.com/sarveshkapre/endpoint-perf-agent/internal/report"
)

type Engine struct {
	detector  *anomaly.Detector
	minRank   int
	cooldown  time.Duration
	lastSent  map[string]time.Time
	prev      *collector.MetricSample
	window    int
	threshold float64
}

func NewEngine(windowSize int, threshold float64, minSeverity string, cooldown time.Duration) (*Engine, error) {
	windowSize, threshold = report.NormalizeParams(windowSize, threshold)

	if minSeverity == "" {
		minSeverity = "low"
	}
	minSeverity = strings.ToLower(strings.TrimSpace(minSeverity))
	minRank, ok := alert.SeverityRank(minSeverity)
	if !ok {
		return nil, fmt.Errorf("unknown severity: %s (expected low|medium|high|critical)", minSeverity)
	}
	if cooldown < 0 {
		return nil, fmt.Errorf("cooldown must be greater than or equal to zero")
	}

	return &Engine{
		detector:  anomaly.NewDetector(windowSize, threshold),
		minRank:   minRank,
		cooldown:  cooldown,
		lastSent:  make(map[string]time.Time),
		window:    windowSize,
		threshold: threshold,
	}, nil
}

func (e *Engine) Params() (windowSize int, threshold float64) {
	return e.window, e.threshold
}

func (e *Engine) Observe(sample collector.MetricSample) []alert.Alert {
	families := collector.DefaultMetricFamilies()
	if sample.MetricFamilies != nil {
		families = *sample.MetricFamilies
	}
	if e.prev == nil {
		// Seed the detector with the absolute metrics so we can start learning immediately.
		if families.CPU {
			_ = e.detector.Check("cpu_percent", sample.CPUPercent)
		}
		if families.Mem {
			_ = e.detector.Check("mem_used_percent", sample.MemUsedPercent)
		}
		if families.Disk {
			_ = e.detector.Check("disk_used_percent", sample.DiskUsedPercent)
		}
		e.prev = &sample
		return nil
	}

	prev := *e.prev
	e.prev = &sample

	prevFamilies := collector.DefaultMetricFamilies()
	if prev.MetricFamilies != nil {
		prevFamilies = *prev.MetricFamilies
	}

	dt := sample.Timestamp.Sub(prev.Timestamp).Seconds()
	if dt <= 0 {
		dt = 1
	}

	metrics := map[string]float64{}
	if families.CPU {
		metrics["cpu_percent"] = sample.CPUPercent
	}
	if families.Mem {
		metrics["mem_used_percent"] = sample.MemUsedPercent
	}
	if families.Disk {
		metrics["disk_used_percent"] = sample.DiskUsedPercent
		if prevFamilies.Disk {
			metrics["disk_read_bytes_per_sec"] = float64(delta(sample.DiskReadBytes, prev.DiskReadBytes)) / dt
			metrics["disk_write_bytes_per_sec"] = float64(delta(sample.DiskWriteBytes, prev.DiskWriteBytes)) / dt
		}
	}
	if families.Net && prevFamilies.Net {
		metrics["net_rx_bytes_per_sec"] = float64(delta(sample.NetRxBytes, prev.NetRxBytes)) / dt
		metrics["net_tx_bytes_per_sec"] = float64(delta(sample.NetTxBytes, prev.NetTxBytes)) / dt
	}

	alerts := make([]alert.Alert, 0)
	for name, value := range metrics {
		a := e.detector.Check(name, value)
		if a == nil {
			continue
		}
		a.Timestamp = sample.Timestamp
		a.TopCPUProcess = toAnomalyProcess(sample.TopCPUProcess)
		a.TopMemProcess = toAnomalyProcess(sample.TopMemProcess)

		rank, ok := alert.SeverityRank(a.Severity)
		if !ok || rank < e.minRank {
			continue
		}
		if e.cooldown > 0 {
			if last, ok := e.lastSent[name]; ok && sample.Timestamp.Sub(last) < e.cooldown {
				continue
			}
			e.lastSent[name] = sample.Timestamp
		}

		alerts = append(alerts, alert.Alert{
			Timestamp:     sample.Timestamp,
			HostID:        sample.HostID,
			Labels:        sample.Labels,
			Metric:        a.Name,
			Value:         a.Value,
			Mean:          a.Mean,
			Stddev:        a.Stddev,
			ZScore:        a.ZScore,
			Severity:      a.Severity,
			Explanation:   a.Explanation,
			TopCPUProcess: a.TopCPUProcess,
			TopMemProcess: a.TopMemProcess,
		})
	}

	return alerts
}

func delta(current, previous uint64) uint64 {
	if current < previous {
		return 0
	}
	return current - previous
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
