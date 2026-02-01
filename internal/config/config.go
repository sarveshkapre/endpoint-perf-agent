package config

import (
	"encoding/json"
	"os"
	"path/filepath"
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
	Interval        time.Duration `json:"-"`
	Duration        time.Duration `json:"-"`
	WindowSize      int           `json:"window_size"`
	ZScoreThreshold float64       `json:"zscore_threshold"`
	OutputPath      string        `json:"output_path"`
	HostID          string        `json:"host_id"`
}

type fileConfig struct {
	Interval        Duration `json:"interval"`
	Duration        Duration `json:"duration"`
	WindowSize      int      `json:"window_size"`
	ZScoreThreshold float64  `json:"zscore_threshold"`
	OutputPath      string   `json:"output_path"`
	HostID          string   `json:"host_id"`
}

func Default() Config {
	return Config{
		Interval:        5 * time.Second,
		Duration:        0,
		WindowSize:      30,
		ZScoreThreshold: 3.0,
		OutputPath:      filepath.Join("data", "metrics.jsonl"),
		HostID:          "",
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
	return cfg, nil
}
