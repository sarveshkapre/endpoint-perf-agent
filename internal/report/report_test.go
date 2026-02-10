package report

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/sarveshkapre/endpoint-perf-agent/internal/collector"
)

func TestAnalyzeDetectsSpike(t *testing.T) {
	start := time.Now().Add(-10 * time.Second).UTC()
	samples := make([]collector.MetricSample, 0, 8)
	cpuBaseline := []float64{10, 11, 9, 10, 12, 11, 10}
	for i := 0; i < len(cpuBaseline); i++ {
		samples = append(samples, collector.MetricSample{
			Timestamp:       start.Add(time.Duration(i) * time.Second),
			CPUPercent:      cpuBaseline[i],
			MemUsedPercent:  40,
			DiskUsedPercent: 50,
			DiskReadBytes:   uint64(i * 100),
			DiskWriteBytes:  uint64(i * 100),
			NetRxBytes:      uint64(i * 100),
			NetTxBytes:      uint64(i * 100),
		})
	}
	samples = append(samples, collector.MetricSample{
		Timestamp:       start.Add(7 * time.Second),
		CPUPercent:      95,
		MemUsedPercent:  40,
		DiskUsedPercent: 50,
		DiskReadBytes:   2000,
		DiskWriteBytes:  2000,
		NetRxBytes:      2000,
		NetTxBytes:      2000,
	})

	result := Analyze(samples, 5, 2.5)
	if result.Samples != len(samples) {
		t.Fatalf("expected %d samples, got %d", len(samples), result.Samples)
	}
	if len(result.Anomalies) == 0 {
		t.Fatal("expected anomaly")
	}
}

func TestAnalyzeSortsSamplesByTimestamp(t *testing.T) {
	t0 := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	samples := []collector.MetricSample{
		{Timestamp: t0.Add(10 * time.Second), CPUPercent: 10},
		{Timestamp: t0, CPUPercent: 10},
	}

	result := Analyze(samples, 5, 3.0)
	if result.Duration <= 0 {
		t.Fatalf("expected positive duration, got %s", result.Duration)
	}
	if !result.FirstTimestamp.Equal(t0) {
		t.Fatalf("expected first timestamp %s, got %s", t0, result.FirstTimestamp)
	}
	if !result.LastTimestamp.Equal(t0.Add(10 * time.Second)) {
		t.Fatalf("expected last timestamp %s, got %s", t0.Add(10*time.Second), result.LastTimestamp)
	}
}

func TestAnalyzeComputesBaselinesAndHostID(t *testing.T) {
	t0 := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	samples := []collector.MetricSample{
		{
			Timestamp:       t0,
			HostID:          "host-01",
			Labels:          map[string]string{"env": "test", "service": "api"},
			CPUPercent:      10,
			MemUsedPercent:  40,
			DiskUsedPercent: 50,
			DiskReadBytes:   0,
		},
		{
			Timestamp:       t0.Add(1 * time.Second),
			HostID:          "host-01",
			Labels:          map[string]string{"env": "test", "service": "api"},
			CPUPercent:      20,
			MemUsedPercent:  40,
			DiskUsedPercent: 50,
			DiskReadBytes:   1000,
		},
		{
			Timestamp:       t0.Add(2 * time.Second),
			HostID:          "host-01",
			Labels:          map[string]string{"env": "test", "service": "api"},
			CPUPercent:      30,
			MemUsedPercent:  40,
			DiskUsedPercent: 50,
			DiskReadBytes:   2000,
		},
	}

	result := Analyze(samples, 5, 3.0)
	if result.HostID != "host-01" {
		t.Fatalf("expected host_id host-01, got %q", result.HostID)
	}
	if got := result.Labels["service"]; got != "api" {
		t.Fatalf("expected stable labels to be detected, got: %+v", result.Labels)
	}

	cpu, ok := result.Baselines["cpu_percent"]
	if !ok {
		t.Fatal("expected cpu_percent baseline")
	}
	if cpu.Count != 3 {
		t.Fatalf("expected cpu_percent count 3, got %d", cpu.Count)
	}
	if cpu.Min != 10 || cpu.Max != 30 {
		t.Fatalf("expected cpu_percent min/max 10/30, got %v/%v", cpu.Min, cpu.Max)
	}
	if cpu.Mean != 20 {
		t.Fatalf("expected cpu_percent mean 20, got %v", cpu.Mean)
	}
	// population stddev for [10,20,30] is sqrt(200/3) ~ 8.1649658
	if cpu.Stddev < 8.16 || cpu.Stddev > 8.17 {
		t.Fatalf("expected cpu_percent stddev ~8.165, got %v", cpu.Stddev)
	}

	diskReadRate, ok := result.Baselines["disk_read_bytes_per_sec"]
	if !ok {
		t.Fatal("expected disk_read_bytes_per_sec baseline")
	}
	if diskReadRate.Count != 2 {
		t.Fatalf("expected disk_read_bytes_per_sec count 2, got %d", diskReadRate.Count)
	}
	if diskReadRate.Mean != 1000 {
		t.Fatalf("expected disk_read_bytes_per_sec mean 1000, got %v", diskReadRate.Mean)
	}

	payload, err := FormatJSON(result)
	if err != nil {
		t.Fatalf("FormatJSON: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if decoded["host_id"] != "host-01" {
		t.Fatalf("expected host_id in json")
	}
	labelsAny, ok := decoded["labels"]
	if !ok {
		t.Fatalf("expected labels in json")
	}
	labels, ok := labelsAny.(map[string]any)
	if !ok || labels["env"] != "test" || labels["service"] != "api" {
		t.Fatalf("unexpected labels in json: %+v", labelsAny)
	}
	if _, ok := decoded["baselines"]; !ok {
		t.Fatalf("expected baselines in json")
	}
}

func TestAnalyzeNormalizesWindowAndThreshold(t *testing.T) {
	t0 := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	samples := []collector.MetricSample{
		{Timestamp: t0, CPUPercent: 10, MemUsedPercent: 40, DiskUsedPercent: 50},
		{Timestamp: t0.Add(1 * time.Second), CPUPercent: 11, MemUsedPercent: 40, DiskUsedPercent: 50},
	}

	result := Analyze(samples, 1, -5)
	if got, want := result.WindowSize, 5; got != want {
		t.Fatalf("expected normalized window %d, got %d", want, got)
	}
	if got, want := result.ZScoreThreshold, 3.0; got != want {
		t.Fatalf("expected normalized threshold %.1f, got %.1f", want, got)
	}
}

func TestAnalyzeAttachesAnomalyProcessContext(t *testing.T) {
	start := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	samples := make([]collector.MetricSample, 0, 8)
	baseline := []float64{10, 11, 9, 10, 12, 11, 10}
	for i := 0; i < len(baseline); i++ {
		samples = append(samples, collector.MetricSample{
			Timestamp:       start.Add(time.Duration(i) * time.Second),
			CPUPercent:      baseline[i],
			MemUsedPercent:  40,
			DiskUsedPercent: 50,
			DiskReadBytes:   uint64(i * 100),
			DiskWriteBytes:  uint64(i * 100),
			NetRxBytes:      uint64(i * 100),
			NetTxBytes:      uint64(i * 100),
		})
	}

	anomalyTimestamp := start.Add(7 * time.Second)
	samples = append(samples, collector.MetricSample{
		Timestamp:       anomalyTimestamp,
		Labels:          map[string]string{"env": "prod"},
		CPUPercent:      95,
		MemUsedPercent:  40,
		DiskUsedPercent: 50,
		DiskReadBytes:   2000,
		DiskWriteBytes:  2000,
		NetRxBytes:      2000,
		NetTxBytes:      2000,
		TopCPUProcess: &collector.ProcessAttribution{
			PID:        1234,
			Name:       "cpu-hog",
			CPUPercent: 88.8,
			RSSBytes:   128 * 1024 * 1024,
		},
		TopMemProcess: &collector.ProcessAttribution{
			PID:        2222,
			Name:       "mem-hog",
			CPUPercent: 12.3,
			RSSBytes:   2 * 1024 * 1024 * 1024,
		},
	})

	result := Analyze(samples, 5, 2.5)
	if len(result.Anomalies) == 0 {
		t.Fatal("expected anomaly")
	}
	first := result.Anomalies[0]
	if first.Timestamp.IsZero() {
		t.Fatal("expected anomaly timestamp")
	}
	if !first.Timestamp.Equal(anomalyTimestamp) {
		t.Fatalf("expected anomaly timestamp %s, got %s", anomalyTimestamp, first.Timestamp)
	}
	if first.TopCPUProcess == nil || first.TopCPUProcess.Name != "cpu-hog" {
		t.Fatalf("expected top cpu process context, got %+v", first.TopCPUProcess)
	}
	if first.TopMemProcess == nil || first.TopMemProcess.Name != "mem-hog" {
		t.Fatalf("expected top memory process context, got %+v", first.TopMemProcess)
	}
	if got := first.Labels["env"]; got != "prod" {
		t.Fatalf("expected anomaly labels to propagate, got: %+v", first.Labels)
	}

	md := FormatMarkdown(result)
	if !strings.Contains(md, "Top CPU process: cpu-hog") {
		t.Fatalf("expected markdown to include top cpu process context, got: %s", md)
	}
	if !strings.Contains(md, "Top memory process: mem-hog") {
		t.Fatalf("expected markdown to include top memory process context, got: %s", md)
	}
}

func TestAnalyze_RespectsMetricFamilies(t *testing.T) {
	t0 := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	families := &collector.MetricFamilies{CPU: true, Mem: true, Disk: false, Net: false}
	samples := []collector.MetricSample{
		{
			Timestamp:      t0,
			CPUPercent:     10,
			MemUsedPercent: 40,
			MetricFamilies: families,
			DiskReadBytes:  1000,
			NetRxBytes:     1000,
		},
		{
			Timestamp:      t0.Add(1 * time.Second),
			CPUPercent:     11,
			MemUsedPercent: 40,
			MetricFamilies: families,
			DiskReadBytes:  2000,
			NetRxBytes:     2000,
		},
	}

	result := Analyze(samples, 5, 3.0)
	if _, ok := result.Baselines["net_rx_bytes_per_sec"]; ok {
		t.Fatalf("expected net_rx_bytes_per_sec to be absent when net family disabled")
	}
	if _, ok := result.Baselines["disk_read_bytes_per_sec"]; ok {
		t.Fatalf("expected disk_read_bytes_per_sec to be absent when disk family disabled")
	}
	if _, ok := result.Baselines["cpu_percent"]; !ok {
		t.Fatalf("expected cpu_percent baseline to exist")
	}
	if _, ok := result.Baselines["mem_used_percent"]; !ok {
		t.Fatalf("expected mem_used_percent baseline to exist")
	}
}
