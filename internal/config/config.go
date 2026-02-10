package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var raw string
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	if raw == "" {
		d.Duration = 0
		return nil
	}
	parsed, err := time.ParseDuration(raw)
	if err != nil {
		return err
	}
	d.Duration = parsed
	return nil
}

type Config struct {
	Interval           time.Duration     `json:"-"`
	Duration           time.Duration     `json:"-"`
	WindowSize         int               `json:"window_size"`
	ZScoreThreshold    float64           `json:"zscore_threshold"`
	OutputPath         string            `json:"output_path"`
	HostID             string            `json:"host_id"`
	Labels             map[string]string `json:"-"`
	ProcessAttribution bool              `json:"process_attribution"`
	Metrics            MetricFamilies    `json:"-"`
}

type fileConfig struct {
	Interval           Duration          `json:"interval"`
	Duration           Duration          `json:"duration"`
	WindowSize         int               `json:"window_size"`
	ZScoreThreshold    float64           `json:"zscore_threshold"`
	OutputPath         string            `json:"output_path"`
	HostID             string            `json:"host_id"`
	Labels             map[string]string `json:"labels"`
	ProcessAttribution *bool             `json:"process_attribution"`
	EnabledMetrics     *[]string         `json:"enabled_metrics"`
}

type MetricFamilies struct {
	CPU  bool
	Mem  bool
	Disk bool
	Net  bool
}

func Default() Config {
	return Config{
		Interval:           5 * time.Second,
		Duration:           0,
		WindowSize:         30,
		ZScoreThreshold:    3.0,
		OutputPath:         filepath.Join("data", "metrics.jsonl"),
		HostID:             "",
		Labels:             nil,
		ProcessAttribution: true,
		Metrics: MetricFamilies{
			CPU:  true,
			Mem:  true,
			Disk: true,
			Net:  true,
		},
	}
}

func Load(path string) (Config, error) {
	cfg := Default()
	if path == "" {
		return cfg, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	var fc fileConfig
	if err := json.Unmarshal(data, &fc); err != nil {
		return cfg, err
	}
	if fc.Interval.Duration != 0 {
		cfg.Interval = fc.Interval.Duration
	}
	if fc.Duration.Duration != 0 {
		cfg.Duration = fc.Duration.Duration
	}
	if fc.WindowSize != 0 {
		cfg.WindowSize = fc.WindowSize
	}
	if fc.ZScoreThreshold != 0 {
		cfg.ZScoreThreshold = fc.ZScoreThreshold
	}
	if fc.OutputPath != "" {
		cfg.OutputPath = fc.OutputPath
	}
	if fc.HostID != "" {
		cfg.HostID = fc.HostID
	}
	if fc.Labels != nil {
		cfg.Labels = fc.Labels
	}
	if fc.ProcessAttribution != nil {
		cfg.ProcessAttribution = *fc.ProcessAttribution
	}
	if fc.EnabledMetrics != nil {
		m, err := ParseMetricFamilies(*fc.EnabledMetrics)
		if err != nil {
			return cfg, err
		}
		cfg.Metrics = m
	}
	return cfg, nil
}

func ParseMetricFamilies(enabled []string) (MetricFamilies, error) {
	if enabled == nil {
		// Field not provided: use defaults.
		return Default().Metrics, nil
	}
	m := MetricFamilies{}
	for _, raw := range enabled {
		name := normalizeMetricName(raw)
		switch name {
		case "cpu":
			m.CPU = true
		case "mem":
			m.Mem = true
		case "disk":
			m.Disk = true
		case "net":
			m.Net = true
		case "":
			// ignore empty entries
		default:
			return MetricFamilies{}, &MetricFamiliesError{Name: raw}
		}
	}
	if !m.Any() {
		return MetricFamilies{}, errors.New("at least one metric family must be enabled")
	}
	return m, nil
}

func (m MetricFamilies) Any() bool {
	return m.CPU || m.Mem || m.Disk || m.Net
}

type MetricFamiliesError struct {
	Name string
}

func (e *MetricFamiliesError) Error() string {
	return "unknown metric family: " + e.Name + " (expected cpu|mem|disk|net)"
}

func normalizeMetricName(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case "memory":
		return "mem"
	case "network":
		return "net"
	default:
		return s
	}
}
