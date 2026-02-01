package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/sarveshkapre/endpoint-perf-agent/internal/collector"
	"github.com/sarveshkapre/endpoint-perf-agent/internal/config"
	"github.com/sarveshkapre/endpoint-perf-agent/internal/report"
	"github.com/sarveshkapre/endpoint-perf-agent/internal/storage"
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
  epagent analyze [flags]
  epagent report [flags]
  epagent version

Commands:
  collect   Sample endpoint metrics and write JSONL to disk.
  analyze   Detect anomalies from collected samples (text or JSON output).
  report    Generate a Markdown report with explanations (use --out - for stdout).
  version   Print the agent version.

Run "epagent <command> -h" for command-specific flags.`)
}

func runCollect(args []string) error {
	fs := flag.NewFlagSet("collect", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	cfgPath := fs.String("config", "", "Path to config file (JSON)")
	interval := fs.Duration("interval", 0, "Sampling interval override (e.g. 2s)")
	duration := fs.Duration("duration", 0, "Total run duration (0 = until interrupted)")
	once := fs.Bool("once", false, "Collect a single sample and exit")
	out := fs.String("out", "", "Output path override for JSONL")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		return err
	}
	if *interval > 0 {
		cfg.Interval = *interval
	}
	if *duration > 0 {
		cfg.Duration = *duration
	}
	if *out != "" {
		cfg.OutputPath = *out
	}
	if *once {
		cfg.Duration = 0
	}

	if cfg.Interval <= 0 {
		return errors.New("interval must be greater than zero")
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

	sampler := collector.NewSampler(cfg.HostID)
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
	format := fs.String("format", "text", "Output format: text|json")
	minSeverity := fs.String("min-severity", "low", "Minimum severity: low|medium|high|critical")
	top := fs.Int("top", 0, "Limit to top N anomalies by absolute z-score (0 = no limit)")
	if err := fs.Parse(args); err != nil {
		return err
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

	result := report.Analyze(samples, cfg.WindowSize, cfg.ZScoreThreshold)
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
	default:
		return fmt.Errorf("unknown format: %s (expected text|json)", *format)
	}
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
	if err := fs.Parse(args); err != nil {
		return err
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

	result := report.Analyze(samples, cfg.WindowSize, cfg.ZScoreThreshold)
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

func exitErr(err error) {
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	os.Exit(1)
}
