package report

import (
	"encoding/json"
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
			CPUPercent:      10,
			MemUsedPercent:  40,
			DiskUsedPercent: 50,
			DiskReadBytes:   0,
		},
		{
			Timestamp:       t0.Add(1 * time.Second),
			HostID:          "host-01",
			CPUPercent:      20,
			MemUsedPercent:  40,
			DiskUsedPercent: 50,
			DiskReadBytes:   1000,
		},
		{
			Timestamp:       t0.Add(2 * time.Second),
			HostID:          "host-01",
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
	if _, ok := decoded["baselines"]; !ok {
		t.Fatalf("expected baselines in json")
	}
}
