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
