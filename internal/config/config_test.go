package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseMetricFamilies(t *testing.T) {
	m, err := ParseMetricFamilies([]string{"cpu", "mem"})
	if err != nil {
		t.Fatalf("ParseMetricFamilies: %v", err)
	}
	if !m.CPU || !m.Mem {
		t.Fatalf("expected cpu/mem enabled, got %+v", m)
	}
	if m.Disk || m.Net {
		t.Fatalf("expected disk/net disabled, got %+v", m)
	}
}

func TestParseMetricFamiliesRejectsUnknown(t *testing.T) {
	_, err := ParseMetricFamilies([]string{"cpu", "nope"})
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestLoadRespectsEnabledMetrics(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cfg.json")
	payload := `{"enabled_metrics":["cpu"]}`
	if err := os.WriteFile(path, []byte(payload), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !cfg.Metrics.CPU || cfg.Metrics.Mem || cfg.Metrics.Disk || cfg.Metrics.Net {
		t.Fatalf("unexpected metrics: %+v", cfg.Metrics)
	}
}
