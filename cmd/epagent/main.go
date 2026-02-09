package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/sarveshkapre/endpoint-perf-agent/internal/alert"
	"github.com/sarveshkapre/endpoint-perf-agent/internal/collector"
	"github.com/sarveshkapre/endpoint-perf-agent/internal/config"
	"github.com/sarveshkapre/endpoint-perf-agent/internal/report"
	"github.com/sarveshkapre/endpoint-perf-agent/internal/storage"
	"github.com/sarveshkapre/endpoint-perf-agent/internal/watch"
)

const version = "0.1.0-dev"

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	switch cmd {
	case "collect":
		if err := runCollect(os.Args[2:]); err != nil {
			exitErr(err)
		}
	case "watch":
		if err := runWatch(os.Args[2:]); err != nil {
			exitErr(err)
		}
	case "analyze":
		if err := runAnalyze(os.Args[2:]); err != nil {
			exitErr(err)
		}
	case "report":
		if err := runReport(os.Args[2:]); err != nil {
			exitErr(err)
		}
	case "version":
		fmt.Println(version)
	case "help", "-h", "--help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Println(`Endpoint Perf Agent

Usage:
  epagent collect [flags]
  epagent watch [flags]
  epagent analyze [flags]
  epagent report [flags]
  epagent version

Commands:
  collect   Sample endpoint metrics and write JSONL to disk.
  watch     Continuously sample and emit anomaly alerts (stdout NDJSON or syslog).
  analyze   Detect anomalies from collected samples (text or JSON output).
  report    Generate a Markdown report with explanations (use --out - for stdout).
  version   Print the agent version.

Run "epagent <command> -h" for command-specific flags.`)
}

func runCollect(args []string) error {
	cfgPath := findFlagStringValue(args, "config")
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return err
	}

	fs := flag.NewFlagSet("collect", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	_ = fs.String("config", cfgPath, "Path to config file (JSON)")
	interval := fs.Duration("interval", cfg.Interval, "Sampling interval (e.g. 2s)")
	duration := fs.Duration("duration", cfg.Duration, "Total run duration (0 = until interrupted)")
	once := fs.Bool("once", false, "Collect a single sample and exit")
	out := fs.String("out", cfg.OutputPath, "Output path for JSONL")
	hostID := fs.String("host-id", "", "Override host ID (defaults to config host_id)")
	metrics := fs.String("metrics", "", "Comma-separated metric families to enable: cpu,mem,disk,net (empty = config/defaults)")
	processAttribution := fs.Bool("process-attribution", cfg.ProcessAttribution, "Capture per-sample top CPU/memory process attribution (can be expensive)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg.Interval = *interval
	cfg.Duration = *duration
	cfg.OutputPath = *out
	cfg.ProcessAttribution = *processAttribution
	if *hostID != "" {
		cfg.HostID = *hostID
	}
	if *metrics != "" {
		m, err := parseMetricFamiliesCSV(*metrics)
		if err != nil {
			return err
		}
		cfg.Metrics = m
	}
	if *once {
		cfg.Duration = 0
	}

	if cfg.Interval <= 0 {
		return errors.New("interval must be greater than zero")
	}
	if cfg.Duration < 0 {
		return errors.New("duration must be greater than or equal to zero")
	}

	if cfg.OutputPath == "" {
		return errors.New("output path is required")
	}

	if err := os.MkdirAll(filepath.Dir(cfg.OutputPath), 0o755); err != nil {
		return err
	}

	writer, err := storage.NewWriter(cfg.OutputPath)
	if err != nil {
		return err
	}
	defer writer.Close()

	sampler := collector.NewSampler(cfg.HostID, cfg.ProcessAttribution, toCollectorMetrics(cfg.Metrics))
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if *once {
		sample, err := sampler.Sample(ctx)
		if err != nil {
			return err
		}
		return writer.Write(sample)
	}

	ticker := time.NewTicker(cfg.Interval)
	defer ticker.Stop()

	deadline := time.Time{}
	if cfg.Duration > 0 {
		deadline = time.Now().Add(cfg.Duration)
	}

	for {
		if !deadline.IsZero() && time.Now().After(deadline) {
			return nil
		}

		sample, err := sampler.Sample(ctx)
		if err != nil {
			return err
		}
		if err := writer.Write(sample); err != nil {
			return err
		}

		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}

func runAnalyze(args []string) error {
	fs := flag.NewFlagSet("analyze", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	in := fs.String("in", "", "Input JSONL path")
	window := fs.Int("window", 0, "Rolling window size override")
	threshold := fs.Float64("threshold", 0, "Z-score threshold override")
	format := fs.String("format", "text", "Output format: text|json|ndjson")
	minSeverity := fs.String("min-severity", "low", "Minimum severity: low|medium|high|critical")
	top := fs.Int("top", 0, "Limit to top N anomalies by absolute z-score (0 = no limit)")
	last := fs.Duration("last", 0, "Analyze only the last duration of samples (relative to the file's last sample timestamp)")
	sinceStr := fs.String("since", "", "Include samples at or after this RFC3339 timestamp (e.g. 2026-02-09T00:00:00Z)")
	untilStr := fs.String("until", "", "Include samples at or before this RFC3339 timestamp (e.g. 2026-02-09T00:01:00Z)")
	sink := fs.String("sink", "stdout", "Alert sink for --format ndjson: stdout|syslog")
	syslogTag := fs.String("syslog-tag", "epagent", "Syslog tag (when --sink syslog)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *window < 0 {
		return errors.New("window must be greater than or equal to zero")
	}
	if *threshold < 0 {
		return errors.New("threshold must be greater than or equal to zero")
	}
	if *top < 0 {
		return errors.New("top must be greater than or equal to zero")
	}
	if *last < 0 {
		return errors.New("last must be greater than or equal to zero")
	}
	if *last > 0 && (*sinceStr != "" || *untilStr != "") {
		return errors.New("cannot combine --last with --since/--until")
	}

	cfg, err := config.Load("")
	if err != nil {
		return err
	}
	if *window > 0 {
		cfg.WindowSize = *window
	}
	if *threshold > 0 {
		cfg.ZScoreThreshold = *threshold
	}

	inputPath := *in
	if inputPath == "" {
		inputPath = cfg.OutputPath
	}
	if inputPath == "" {
		return errors.New("input path is required")
	}

	samples, err := storage.ReadSamples(inputPath)
	if err != nil {
		return err
	}

	if *last > 0 && len(samples) > 0 {
		maxTS := samples[0].Timestamp
		for _, s := range samples[1:] {
			if s.Timestamp.After(maxTS) {
				maxTS = s.Timestamp
			}
		}
		since := maxTS.Add(-*last)
		samples, err = report.FilterSamplesByTime(samples, since, maxTS)
		if err != nil {
			return err
		}
	} else {
		since, err := parseRFC3339TimeFlag("since", *sinceStr)
		if err != nil {
			return err
		}
		until, err := parseRFC3339TimeFlag("until", *untilStr)
		if err != nil {
			return err
		}
		samples, err = report.FilterSamplesByTime(samples, since, until)
		if err != nil {
			return err
		}
	}

	windowSize, zScoreThreshold := report.NormalizeParams(cfg.WindowSize, cfg.ZScoreThreshold)
	result := report.Analyze(samples, windowSize, zScoreThreshold)
	result, err = report.ApplyFilters(result, *minSeverity, *top)
	if err != nil {
		return err
	}
	switch *format {
	case "text":
		fmt.Println(report.FormatSummary(result))
		return nil
	case "json":
		payload, err := report.FormatJSON(result)
		if err != nil {
			return err
		}
		fmt.Println(string(payload))
		return nil
	case "ndjson":
		var alertSink alert.Sink
		switch *sink {
		case "stdout":
			alertSink = alert.NewStdoutSink(os.Stdout)
		case "syslog":
			s, err := alert.NewSyslogSink(*syslogTag)
			if err != nil {
				return err
			}
			alertSink = s
			defer alertSink.Close()
		default:
			return fmt.Errorf("unknown sink: %s (expected stdout|syslog)", *sink)
		}

		for _, a := range result.Anomalies {
			if err := alertSink.Emit(context.Background(), alert.Alert{
				Timestamp:     a.Timestamp,
				HostID:        result.HostID,
				Metric:        a.Name,
				Value:         a.Value,
				Mean:          a.Mean,
				Stddev:        a.Stddev,
				ZScore:        a.ZScore,
				Severity:      a.Severity,
				Explanation:   a.Explanation,
				TopCPUProcess: a.TopCPUProcess,
				TopMemProcess: a.TopMemProcess,
			}); err != nil {
				return err
			}
		}
		return nil
	default:
		return fmt.Errorf("unknown format: %s (expected text|json|ndjson)", *format)
	}
}

func runWatch(args []string) error {
	cfgPath := findFlagStringValue(args, "config")
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return err
	}

	fs := flag.NewFlagSet("watch", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	_ = fs.String("config", cfgPath, "Path to config file (JSON)")
	interval := fs.Duration("interval", cfg.Interval, "Sampling interval (e.g. 2s)")
	duration := fs.Duration("duration", cfg.Duration, "Total run duration (0 = until interrupted)")
	out := fs.String("out", "", "Optional JSONL path to also write samples (empty = don't write)")
	hostID := fs.String("host-id", "", "Override host ID (defaults to config host_id)")
	window := fs.Int("window", cfg.WindowSize, "Rolling window size")
	threshold := fs.Float64("threshold", cfg.ZScoreThreshold, "Z-score threshold")
	minSeverity := fs.String("min-severity", "medium", "Minimum severity to emit: low|medium|high|critical")
	sink := fs.String("sink", "stdout", "Alert sink: stdout|syslog")
	syslogTag := fs.String("syslog-tag", "epagent", "Syslog tag (when --sink syslog)")
	cooldown := fs.Duration("cooldown", 30*time.Second, "Per-metric alert cooldown (0 = no dedupe)")
	metrics := fs.String("metrics", "", "Comma-separated metric families to enable: cpu,mem,disk,net (empty = config/defaults)")
	processAttribution := fs.Bool("process-attribution", cfg.ProcessAttribution, "Capture per-sample top CPU/memory process attribution (can be expensive)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *interval < 0 {
		return errors.New("interval must be greater than or equal to zero")
	}
	if *duration < 0 {
		return errors.New("duration must be greater than or equal to zero")
	}
	if *window < 0 {
		return errors.New("window must be greater than or equal to zero")
	}
	if *threshold < 0 {
		return errors.New("threshold must be greater than or equal to zero")
	}
	if *cooldown < 0 {
		return errors.New("cooldown must be greater than or equal to zero")
	}

	cfg.Interval = *interval
	cfg.Duration = *duration
	cfg.WindowSize = *window
	cfg.ZScoreThreshold = *threshold
	cfg.ProcessAttribution = *processAttribution
	if *hostID != "" {
		cfg.HostID = *hostID
	}
	if *metrics != "" {
		m, err := parseMetricFamiliesCSV(*metrics)
		if err != nil {
			return err
		}
		cfg.Metrics = m
	}

	if cfg.Interval <= 0 {
		return errors.New("interval must be greater than zero")
	}

	var writer watch.SampleWriter
	var writerCloser interface{ Close() error }
	if *out != "" {
		if err := os.MkdirAll(filepath.Dir(*out), 0o755); err != nil {
			return err
		}
		w, err := storage.NewWriter(*out)
		if err != nil {
			return err
		}
		writer = w
		writerCloser = w
		defer writerCloser.Close()
	}

	sampler := collector.NewSampler(cfg.HostID, cfg.ProcessAttribution, toCollectorMetrics(cfg.Metrics))

	engine, err := watch.NewEngine(cfg.WindowSize, cfg.ZScoreThreshold, *minSeverity, *cooldown)
	if err != nil {
		return err
	}

	var alertSink alert.Sink
	switch *sink {
	case "stdout":
		alertSink = alert.NewStdoutSink(os.Stdout)
	case "syslog":
		s, err := alert.NewSyslogSink(*syslogTag)
		if err != nil {
			return err
		}
		alertSink = s
		defer alertSink.Close()
	default:
		return fmt.Errorf("unknown sink: %s (expected stdout|syslog)", *sink)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	runner := &watch.Runner{
		Sampler:  sampler,
		Engine:   engine,
		Sink:     alertSink,
		Interval: cfg.Interval,
		Duration: cfg.Duration,
		Writer:   writer,
	}
	return runner.Run(ctx)
}

func parseMetricFamiliesCSV(csv string) (config.MetricFamilies, error) {
	parts := strings.Split(csv, ",")
	enabled := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		enabled = append(enabled, p)
	}
	return config.ParseMetricFamilies(enabled)
}

func toCollectorMetrics(m config.MetricFamilies) collector.MetricFamilies {
	return collector.MetricFamilies{CPU: m.CPU, Mem: m.Mem, Disk: m.Disk, Net: m.Net}
}

func runReport(args []string) error {
	fs := flag.NewFlagSet("report", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	in := fs.String("in", "", "Input JSONL path")
	out := fs.String("out", "endpoint-perf-report.md", "Output markdown path")
	window := fs.Int("window", 0, "Rolling window size override")
	threshold := fs.Float64("threshold", 0, "Z-score threshold override")
	minSeverity := fs.String("min-severity", "low", "Minimum severity: low|medium|high|critical")
	top := fs.Int("top", 0, "Limit to top N anomalies by absolute z-score (0 = no limit)")
	last := fs.Duration("last", 0, "Report only the last duration of samples (relative to the file's last sample timestamp)")
	sinceStr := fs.String("since", "", "Include samples at or after this RFC3339 timestamp (e.g. 2026-02-09T00:00:00Z)")
	untilStr := fs.String("until", "", "Include samples at or before this RFC3339 timestamp (e.g. 2026-02-09T00:01:00Z)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *window < 0 {
		return errors.New("window must be greater than or equal to zero")
	}
	if *threshold < 0 {
		return errors.New("threshold must be greater than or equal to zero")
	}
	if *top < 0 {
		return errors.New("top must be greater than or equal to zero")
	}
	if *last < 0 {
		return errors.New("last must be greater than or equal to zero")
	}
	if *last > 0 && (*sinceStr != "" || *untilStr != "") {
		return errors.New("cannot combine --last with --since/--until")
	}

	cfg, err := config.Load("")
	if err != nil {
		return err
	}
	if *window > 0 {
		cfg.WindowSize = *window
	}
	if *threshold > 0 {
		cfg.ZScoreThreshold = *threshold
	}

	inputPath := *in
	if inputPath == "" {
		inputPath = cfg.OutputPath
	}
	if inputPath == "" {
		return errors.New("input path is required")
	}

	samples, err := storage.ReadSamples(inputPath)
	if err != nil {
		return err
	}

	if *last > 0 && len(samples) > 0 {
		maxTS := samples[0].Timestamp
		for _, s := range samples[1:] {
			if s.Timestamp.After(maxTS) {
				maxTS = s.Timestamp
			}
		}
		since := maxTS.Add(-*last)
		samples, err = report.FilterSamplesByTime(samples, since, maxTS)
		if err != nil {
			return err
		}
	} else {
		since, err := parseRFC3339TimeFlag("since", *sinceStr)
		if err != nil {
			return err
		}
		until, err := parseRFC3339TimeFlag("until", *untilStr)
		if err != nil {
			return err
		}
		samples, err = report.FilterSamplesByTime(samples, since, until)
		if err != nil {
			return err
		}
	}

	windowSize, zScoreThreshold := report.NormalizeParams(cfg.WindowSize, cfg.ZScoreThreshold)
	result := report.Analyze(samples, windowSize, zScoreThreshold)
	result, err = report.ApplyFilters(result, *minSeverity, *top)
	if err != nil {
		return err
	}
	md := report.FormatMarkdown(result)
	if *out == "-" {
		fmt.Print(md)
		return nil
	}
	if err := os.WriteFile(*out, []byte(md), 0o644); err != nil {
		return err
	}
	fmt.Printf("report written to %s\n", *out)
	return nil
}

func findFlagStringValue(args []string, name string) string {
	prefix1 := "-" + name + "="
	prefix2 := "--" + name + "="
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "-"+name || a == "--"+name:
			if i+1 < len(args) {
				return args[i+1]
			}
		case strings.HasPrefix(a, prefix1):
			return strings.TrimPrefix(a, prefix1)
		case strings.HasPrefix(a, prefix2):
			return strings.TrimPrefix(a, prefix2)
		}
	}
	return ""
}

func parseRFC3339TimeFlag(name, value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, nil
	}
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("%s must be an RFC3339 timestamp (e.g. 2026-02-09T00:00:00Z): %w", name, err)
	}
	return t, nil
}

func exitErr(err error) {
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	os.Exit(1)
}
