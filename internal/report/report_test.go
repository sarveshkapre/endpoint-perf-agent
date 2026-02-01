package report

import (
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
