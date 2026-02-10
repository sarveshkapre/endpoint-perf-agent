package selftest

import (
	"context"
	"runtime"
	"sort"
	"time"

	"github.com/sarveshkapre/endpoint-perf-agent/internal/collector"
	"github.com/shirou/gopsutil/v3/process"
)

type Check struct {
	Name       string        `json:"name"`
	OK         bool          `json:"ok"`
	Error      string        `json:"error,omitempty"`
	Runs       int           `json:"runs"`
	MedianTime time.Duration `json:"median_time"`
	P95Time    time.Duration `json:"p95_time"`
}

type Result struct {
	Timestamp time.Time `json:"timestamp"`
	GOOS      string    `json:"goos"`
	GOARCH    string    `json:"goarch"`

	EnabledMetrics     []string `json:"enabled_metrics"`
	ProcessAttribution bool     `json:"process_attribution"`

	ProcessListOK    bool   `json:"process_list_ok"`
	ProcessListError string `json:"process_list_error,omitempty"`
	ProcessCount     int    `json:"process_count,omitempty"`

	Checks []Check `json:"checks"`
}

type Options struct {
	Metrics            collector.MetricFamilies
	ProcessAttribution bool
	Runs               int
	TimeoutPerRun      time.Duration
}

func Run(ctx context.Context, opts Options) Result {
	if opts.Runs <= 0 {
		opts.Runs = 3
	}
	if opts.TimeoutPerRun <= 0 {
		opts.TimeoutPerRun = 2 * time.Second
	}

	res := Result{
		Timestamp:          time.Now().UTC(),
		GOOS:               runtime.GOOS,
		GOARCH:             runtime.GOARCH,
		EnabledMetrics:     enabledMetricNames(opts.Metrics),
		ProcessAttribution: opts.ProcessAttribution,
	}

	// Process list access check (best-effort; process attribution can still partially work).
	{
		cctx, cancel := context.WithTimeout(ctx, opts.TimeoutPerRun)
		defer cancel()
		procs, err := process.ProcessesWithContext(cctx)
		if err != nil {
			res.ProcessListOK = false
			res.ProcessListError = err.Error()
		} else {
			res.ProcessListOK = true
			res.ProcessCount = len(procs)
		}
	}

	// Per-family checks to isolate failures.
	for _, fam := range []struct {
		name    string
		metrics collector.MetricFamilies
	}{
		{name: "cpu", metrics: collector.MetricFamilies{CPU: true}},
		{name: "mem", metrics: collector.MetricFamilies{Mem: true}},
		{name: "disk", metrics: collector.MetricFamilies{Disk: true}},
		{name: "net", metrics: collector.MetricFamilies{Net: true}},
	} {
		if !isFamilyEnabled(opts.Metrics, fam.name) {
			continue
		}
		res.Checks = append(res.Checks, measureSampler(ctx, fam.name, opts.Runs, opts.TimeoutPerRun, fam.metrics, false))
	}

	// Combined check: baseline (no process attribution).
	res.Checks = append(res.Checks, measureSampler(ctx, "combined", opts.Runs, opts.TimeoutPerRun, opts.Metrics, false))

	// Combined check: with process attribution enabled (if requested).
	if opts.ProcessAttribution {
		res.Checks = append(res.Checks, measureSampler(ctx, "combined+process", opts.Runs, opts.TimeoutPerRun, opts.Metrics, true))
	}

	return res
}

func measureSampler(ctx context.Context, name string, runs int, timeout time.Duration, metrics collector.MetricFamilies, processAttribution bool) Check {
	s := collector.NewSampler("", nil, processAttribution, metrics)

	durations := make([]time.Duration, 0, runs)
	for i := 0; i < runs; i++ {
		cctx, cancel := context.WithTimeout(ctx, timeout)
		start := time.Now()
		_, err := s.Sample(cctx)
		cancel()
		if err != nil {
			return Check{Name: name, OK: false, Error: err.Error(), Runs: i + 1}
		}
		durations = append(durations, time.Since(start))
	}

	sort.Slice(durations, func(i, j int) bool { return durations[i] < durations[j] })
	return Check{
		Name:       name,
		OK:         true,
		Runs:       runs,
		MedianTime: percentile(durations, 0.50),
		P95Time:    percentile(durations, 0.95),
	}
}

func percentile(sorted []time.Duration, p float64) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	if p <= 0 {
		return sorted[0]
	}
	if p >= 1 {
		return sorted[len(sorted)-1]
	}
	pos := int(float64(len(sorted)-1)*p + 0.5)
	if pos < 0 {
		pos = 0
	}
	if pos >= len(sorted) {
		pos = len(sorted) - 1
	}
	return sorted[pos]
}

func enabledMetricNames(m collector.MetricFamilies) []string {
	out := make([]string, 0, 4)
	if m.CPU {
		out = append(out, "cpu")
	}
	if m.Mem {
		out = append(out, "mem")
	}
	if m.Disk {
		out = append(out, "disk")
	}
	if m.Net {
		out = append(out, "net")
	}
	return out
}

func isFamilyEnabled(m collector.MetricFamilies, name string) bool {
	switch name {
	case "cpu":
		return m.CPU
	case "mem":
		return m.Mem
	case "disk":
		return m.Disk
	case "net":
		return m.Net
	default:
		return false
	}
}
