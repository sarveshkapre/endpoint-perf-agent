package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sarveshkapre/endpoint-perf-agent/internal/anomaly"
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

func TestLoadRespectsLabels(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cfg.json")
	payload := `{"labels":{"env":"test","service":"api"}}`
	if err := os.WriteFile(path, []byte(payload), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got := cfg.Labels["env"]; got != "test" {
		t.Fatalf("expected env label, got %+v", cfg.Labels)
	}
	if got := cfg.Labels["service"]; got != "api" {
		t.Fatalf("expected service label, got %+v", cfg.Labels)
	}
}

func TestParseStaticThresholdsNormalizesAliases(t *testing.T) {
	thresholds, err := ParseStaticThresholds(map[string]float64{
		"cpu":        80,
		"disk_read":  1024,
		"net_rx_bps": 2048,
	})
	if err != nil {
		t.Fatalf("ParseStaticThresholds: %v", err)
	}
	if got := thresholds["cpu_percent"]; got != 80 {
		t.Fatalf("expected cpu_percent threshold, got %+v", thresholds)
	}
	if got := thresholds["disk_read_bytes_per_sec"]; got != 1024 {
		t.Fatalf("expected disk_read_bytes_per_sec threshold, got %+v", thresholds)
	}
	if got := thresholds["net_rx_bytes_per_sec"]; got != 2048 {
		t.Fatalf("expected net_rx_bytes_per_sec threshold, got %+v", thresholds)
	}
}

func TestParseStaticThresholdsRejectsUnknownMetric(t *testing.T) {
	_, err := ParseStaticThresholds(map[string]float64{"nope": 1})
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestParseStaticThresholdsRejectsNonPositive(t *testing.T) {
	_, err := ParseStaticThresholds(map[string]float64{"cpu": 0})
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestLoadRespectsStaticThresholds(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cfg.json")
	payload := `{"static_thresholds":{"cpu":85,"net_tx_bytes_per_sec":4096}}`
	if err := os.WriteFile(path, []byte(payload), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got := cfg.StaticThresholds["cpu_percent"]; got != 85 {
		t.Fatalf("expected cpu_percent threshold 85, got %+v", cfg.StaticThresholds)
	}
	if got := cfg.StaticThresholds["net_tx_bytes_per_sec"]; got != 4096 {
		t.Fatalf("expected net_tx_bytes_per_sec threshold 4096, got %+v", cfg.StaticThresholds)
	}
}

func TestLoadRespectsSamplingJitter(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cfg.json")
	payload := `{"sampling_jitter":"750ms"}`
	if err := os.WriteFile(path, []byte(payload), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got, want := cfg.SamplingJitter.String(), "750ms"; got != want {
		t.Fatalf("expected sampling jitter %q, got %q", want, got)
	}
}

func TestParseCooldownOverridesNormalizesAliases(t *testing.T) {
	cooldowns, err := ParseCooldownOverrides(map[string]time.Duration{
		"cpu":       5 * time.Second,
		"disk_read": 10 * time.Second,
	})
	if err != nil {
		t.Fatalf("ParseCooldownOverrides: %v", err)
	}
	if got := cooldowns["cpu_percent"]; got != 5*time.Second {
		t.Fatalf("expected cpu_percent cooldown, got %+v", cooldowns)
	}
	if got := cooldowns["disk_read_bytes_per_sec"]; got != 10*time.Second {
		t.Fatalf("expected disk_read_bytes_per_sec cooldown, got %+v", cooldowns)
	}
}

func TestParseCooldownOverridesRejectsUnknownMetric(t *testing.T) {
	_, err := ParseCooldownOverrides(map[string]time.Duration{"nope": time.Second})
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestLoadRespectsCooldownOverrides(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cfg.json")
	payload := `{"cooldowns":{"cpu":"5s","net_rx_bytes_per_sec":"2s"}}`
	if err := os.WriteFile(path, []byte(payload), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got := cfg.CooldownOverrides["cpu_percent"]; got != 5*time.Second {
		t.Fatalf("expected cpu_percent cooldown 5s, got %+v", cfg.CooldownOverrides)
	}
	if got := cfg.CooldownOverrides["net_rx_bytes_per_sec"]; got != 2*time.Second {
		t.Fatalf("expected net_rx_bytes_per_sec cooldown 2s, got %+v", cfg.CooldownOverrides)
	}
}

func TestParsePercentileRulesNormalizesAliases(t *testing.T) {
	rules, err := ParsePercentileRules(map[string]anomaly.PercentileRule{
		"cpu": {
			Percentile: 95,
			Multiplier: 1.2,
		},
	})
	if err != nil {
		t.Fatalf("ParsePercentileRules: %v", err)
	}
	rule, ok := rules["cpu_percent"]
	if !ok {
		t.Fatalf("expected cpu_percent rule, got %+v", rules)
	}
	if rule.Percentile != 95 || rule.Multiplier != 1.2 {
		t.Fatalf("unexpected rule values: %+v", rule)
	}
}

func TestParsePercentileRulesRejectsUnknownMetric(t *testing.T) {
	_, err := ParsePercentileRules(map[string]anomaly.PercentileRule{
		"nope": {Percentile: 95, Multiplier: 1.2},
	})
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestLoadRespectsPercentileRules(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cfg.json")
	payload := `{"percentile_thresholds":{"cpu":{"percentile":95,"multiplier":1.2}}}`
	if err := os.WriteFile(path, []byte(payload), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	rule, ok := cfg.PercentileRules["cpu_percent"]
	if !ok {
		t.Fatalf("expected cpu_percent percentile rule, got %+v", cfg.PercentileRules)
	}
	if rule.Percentile != 95 || rule.Multiplier != 1.2 {
		t.Fatalf("unexpected percentile rule: %+v", rule)
	}
}
