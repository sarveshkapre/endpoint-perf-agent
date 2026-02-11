package watch

import (
	"testing"
	"time"

	"github.com/sarveshkapre/endpoint-perf-agent/internal/collector"
)

func TestEngine_EmitsAlert(t *testing.T) {
	engine, err := NewEngine(5, 3.0, nil, "low", 0)
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}

	base := time.Date(2026, 2, 9, 0, 0, 0, 0, time.UTC)
	values := []float64{10, 11, 9, 10, 12, 50}
	var gotCPU bool

	for i, v := range values {
		s := collector.MetricSample{
			Timestamp:       base.Add(time.Duration(i) * time.Second),
			HostID:          "host-1",
			Labels:          map[string]string{"env": "test"},
			CPUPercent:      v,
			MemUsedPercent:  20,
			DiskUsedPercent: 30,
			DiskReadBytes:   uint64(i * 100),
			DiskWriteBytes:  uint64(i * 200),
			NetRxBytes:      uint64(i * 300),
			NetTxBytes:      uint64(i * 400),
		}

		for _, a := range engine.Observe(s) {
			if a.Metric == "cpu_percent" {
				gotCPU = true
				if a.HostID != "host-1" {
					t.Fatalf("unexpected host_id: %q", a.HostID)
				}
				if got := a.Labels["env"]; got != "test" {
					t.Fatalf("expected labels to propagate, got: %+v", a.Labels)
				}
				if a.Timestamp.IsZero() {
					t.Fatalf("expected timestamp")
				}
				if a.Severity == "" {
					t.Fatalf("expected severity")
				}
			}
		}
	}

	if !gotCPU {
		t.Fatalf("expected at least one cpu_percent alert")
	}
}

func TestEngine_CooldownSuppressesDuplicates(t *testing.T) {
	engine, err := NewEngine(5, 3.0, nil, "low", time.Minute)
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}

	base := time.Date(2026, 2, 9, 0, 0, 0, 0, time.UTC)
	// Two spikes within cooldown should only emit once for cpu_percent.
	values := []float64{10, 11, 9, 10, 12, 50, 80}

	cpuAlerts := 0
	for i, v := range values {
		s := collector.MetricSample{
			Timestamp:       base.Add(time.Duration(i) * 10 * time.Second),
			HostID:          "host-1",
			CPUPercent:      v,
			MemUsedPercent:  20,
			DiskUsedPercent: 30,
			DiskReadBytes:   uint64(i * 100),
			DiskWriteBytes:  uint64(i * 200),
			NetRxBytes:      uint64(i * 300),
			NetTxBytes:      uint64(i * 400),
		}

		for _, a := range engine.Observe(s) {
			if a.Metric == "cpu_percent" {
				cpuAlerts++
			}
		}
	}

	if cpuAlerts != 1 {
		t.Fatalf("expected 1 cpu_percent alert due to cooldown; got %d", cpuAlerts)
	}
}

func TestNewEngine_RejectsUnknownSeverity(t *testing.T) {
	_, err := NewEngine(5, 3.0, nil, "nope", 0)
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestEngine_RespectsMetricFamilies(t *testing.T) {
	engine, err := NewEngine(5, 3.0, nil, "low", 0)
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}

	base := time.Date(2026, 2, 9, 0, 0, 0, 0, time.UTC)
	values := []float64{10, 11, 9, 10, 12, 50}
	disabledCPU := &collector.MetricFamilies{CPU: false, Mem: true, Disk: true, Net: true}

	var gotCPU bool
	for i, v := range values {
		s := collector.MetricSample{
			Timestamp:      base.Add(time.Duration(i) * time.Second),
			HostID:         "host-1",
			CPUPercent:     v,
			MemUsedPercent: 20,
			MetricFamilies: disabledCPU,
		}
		for _, a := range engine.Observe(s) {
			if a.Metric == "cpu_percent" {
				gotCPU = true
			}
		}
	}
	if gotCPU {
		t.Fatalf("expected no cpu_percent alerts when CPU family is disabled")
	}
}

func TestEngine_EmitsStaticThresholdAlert(t *testing.T) {
	engine, err := NewEngine(5, 10.0, map[string]float64{"cpu_percent": 50}, "low", 0)
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}

	base := time.Date(2026, 2, 9, 0, 0, 0, 0, time.UTC)
	_ = engine.Observe(collector.MetricSample{
		Timestamp:      base,
		CPUPercent:     10,
		MemUsedPercent: 20,
	})
	alerts := engine.Observe(collector.MetricSample{
		Timestamp:      base.Add(time.Second),
		CPUPercent:     70,
		MemUsedPercent: 20,
	})
	if len(alerts) == 0 {
		t.Fatalf("expected static-threshold alert")
	}
	if alerts[0].RuleType != "static_threshold" {
		t.Fatalf("expected static_threshold rule type, got %q", alerts[0].RuleType)
	}
	if alerts[0].Threshold != 50 {
		t.Fatalf("expected threshold 50, got %v", alerts[0].Threshold)
	}
}
