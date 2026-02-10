package report

import (
	"testing"

	"github.com/sarveshkapre/endpoint-perf-agent/internal/anomaly"
	"github.com/sarveshkapre/endpoint-perf-agent/internal/collector"
)

func TestFilterByMetricFamilies_FiltersBaselinesAndAnomalies(t *testing.T) {
	in := AnalysisResult{
		Anomalies: []anomaly.Anomaly{
			{Name: "cpu_percent"},
			{Name: "mem_used_percent"},
			{Name: "net_rx_bytes_per_sec"},
			{Name: "disk_read_bytes_per_sec"},
		},
		Baselines: map[string]MetricStats{
			"cpu_percent":              {Count: 1, Mean: 1},
			"mem_used_percent":         {Count: 1, Mean: 1},
			"net_rx_bytes_per_sec":     {Count: 1, Mean: 1},
			"disk_read_bytes_per_sec":  {Count: 1, Mean: 1},
			"disk_write_bytes_per_sec": {Count: 1, Mean: 1},
		},
	}

	out := FilterByMetricFamilies(in, collector.MetricFamilies{CPU: true, Mem: false, Disk: false, Net: true})

	if _, ok := out.Baselines["cpu_percent"]; !ok {
		t.Fatalf("expected cpu_percent baseline to remain")
	}
	if _, ok := out.Baselines["net_rx_bytes_per_sec"]; !ok {
		t.Fatalf("expected net_rx_bytes_per_sec baseline to remain")
	}
	if _, ok := out.Baselines["mem_used_percent"]; ok {
		t.Fatalf("expected mem_used_percent baseline to be filtered")
	}
	if _, ok := out.Baselines["disk_read_bytes_per_sec"]; ok {
		t.Fatalf("expected disk_read_bytes_per_sec baseline to be filtered")
	}

	names := map[string]bool{}
	for _, a := range out.Anomalies {
		names[a.Name] = true
	}
	if !names["cpu_percent"] || !names["net_rx_bytes_per_sec"] {
		t.Fatalf("expected cpu and net anomalies to remain, got: %+v", out.Anomalies)
	}
	if names["mem_used_percent"] || names["disk_read_bytes_per_sec"] {
		t.Fatalf("expected mem and disk anomalies to be filtered, got: %+v", out.Anomalies)
	}
}
