package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sarveshkapre/endpoint-perf-agent/internal/anomaly"
)

func writeSamplesJSONL(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "metrics.jsonl")
	data := "" +
		`{"timestamp":"2026-02-09T00:00:00Z","host_id":"test","cpu_percent":10,"mem_used_percent":20,"disk_used_percent":30,"disk_read_bytes":100,"disk_write_bytes":200,"net_rx_bytes":300,"net_tx_bytes":400}` + "\n" +
		`{"timestamp":"2026-02-09T00:00:01Z","host_id":"test","cpu_percent":11,"mem_used_percent":20,"disk_used_percent":30,"disk_read_bytes":150,"disk_write_bytes":250,"net_rx_bytes":330,"net_tx_bytes":450}` + "\n" +
		`{"timestamp":"2026-02-09T00:00:02Z","host_id":"test","cpu_percent":9,"mem_used_percent":20,"disk_used_percent":30,"disk_read_bytes":190,"disk_write_bytes":260,"net_rx_bytes":360,"net_tx_bytes":470}` + "\n" +
		`{"timestamp":"2026-02-09T00:00:03Z","host_id":"test","cpu_percent":10,"mem_used_percent":21,"disk_used_percent":30,"disk_read_bytes":220,"disk_write_bytes":270,"net_rx_bytes":390,"net_tx_bytes":490}` + "\n" +
		`{"timestamp":"2026-02-09T00:00:04Z","host_id":"test","cpu_percent":12,"mem_used_percent":21,"disk_used_percent":30,"disk_read_bytes":260,"disk_write_bytes":280,"net_rx_bytes":420,"net_tx_bytes":510}` + "\n"
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatalf("write samples: %v", err)
	}
	return path
}

func TestAnalyze_RejectsUnknownSeverity(t *testing.T) {
	in := writeSamplesJSONL(t)
	if err := runAnalyze([]string{"--in", in, "--min-severity", "nope"}); err == nil {
		t.Fatalf("expected error")
	}
}

func TestAnalyze_JSONOutput(t *testing.T) {
	in := writeSamplesJSONL(t)
	if err := runAnalyze([]string{"--in", in, "--format", "json", "--window", "5", "--threshold", "3"}); err != nil {
		t.Fatalf("runAnalyze: %v", err)
	}
}

func TestAnalyze_NDJSONOutput(t *testing.T) {
	in := writeSamplesJSONL(t)
	if err := runAnalyze([]string{"--in", in, "--format", "ndjson", "--sink", "stdout", "--window", "5", "--threshold", "3"}); err != nil {
		t.Fatalf("runAnalyze: %v", err)
	}
}

func TestAnalyze_RejectsUnknownSinkForNDJSON(t *testing.T) {
	in := writeSamplesJSONL(t)
	if err := runAnalyze([]string{"--in", in, "--format", "ndjson", "--sink", "nope"}); err == nil {
		t.Fatalf("expected error")
	}
}

func TestAnalyze_RejectsUnknownMetricFamilyFilter(t *testing.T) {
	in := writeSamplesJSONL(t)
	if err := runAnalyze([]string{"--in", in, "--metric", "nope"}); err == nil {
		t.Fatalf("expected error")
	}
}

func TestAnalyze_RejectsUnknownRedactMode(t *testing.T) {
	in := writeSamplesJSONL(t)
	if err := runAnalyze([]string{"--in", in, "--redact", "nope"}); err == nil {
		t.Fatalf("expected error")
	}
}

func TestAnalyze_RejectsUnknownStaticThresholdMetric(t *testing.T) {
	in := writeSamplesJSONL(t)
	if err := runAnalyze([]string{"--in", in, "--static-threshold", "nope=1"}); err == nil {
		t.Fatalf("expected error")
	}
}

func TestAnalyze_RejectsInvalidPercentileThreshold(t *testing.T) {
	in := writeSamplesJSONL(t)
	if err := runAnalyze([]string{"--in", in, "--percentile-threshold", "cpu=0,1.2"}); err == nil {
		t.Fatalf("expected error")
	}
}

func TestAnalyze_RejectsInvalidSince(t *testing.T) {
	in := writeSamplesJSONL(t)
	if err := runAnalyze([]string{"--in", in, "--since", "nope"}); err == nil {
		t.Fatalf("expected error")
	}
}

func TestAnalyze_RejectsSinceAfterUntil(t *testing.T) {
	in := writeSamplesJSONL(t)
	if err := runAnalyze([]string{"--in", in, "--since", "2026-02-09T00:00:03Z", "--until", "2026-02-09T00:00:02Z"}); err == nil {
		t.Fatalf("expected error")
	}
}

func TestAnalyze_AcceptsFractionalRFC3339Since(t *testing.T) {
	in := writeSamplesJSONL(t)
	if err := runAnalyze([]string{"--in", in, "--format", "json", "--window", "5", "--threshold", "3", "--since", "2026-02-09T00:00:01.123Z"}); err != nil {
		t.Fatalf("runAnalyze: %v", err)
	}
}

func TestAnalyze_LastRejectsNegative(t *testing.T) {
	in := writeSamplesJSONL(t)
	if err := runAnalyze([]string{"--in", in, "--last", "-1s"}); err == nil {
		t.Fatalf("expected error")
	}
}

func TestAnalyze_LastCannotCombineSince(t *testing.T) {
	in := writeSamplesJSONL(t)
	if err := runAnalyze([]string{"--in", in, "--last", "1s", "--since", "2026-02-09T00:00:01Z"}); err == nil {
		t.Fatalf("expected error")
	}
}

func TestAnalyze_LastWorks(t *testing.T) {
	in := writeSamplesJSONL(t)
	if err := runAnalyze([]string{"--in", in, "--format", "json", "--window", "5", "--threshold", "3", "--last", "2s"}); err != nil {
		t.Fatalf("runAnalyze: %v", err)
	}
}

func TestReport_WritesToStdout(t *testing.T) {
	in := writeSamplesJSONL(t)
	if err := runReport([]string{"--in", in, "--out", "-", "--window", "5", "--threshold", "3"}); err != nil {
		t.Fatalf("runReport: %v", err)
	}
}

func TestReport_RejectsInvalidUntil(t *testing.T) {
	in := writeSamplesJSONL(t)
	if err := runReport([]string{"--in", in, "--out", "-", "--until", "nope"}); err == nil {
		t.Fatalf("expected error")
	}
}

func TestReport_RejectsUnknownMetricFamilyFilter(t *testing.T) {
	in := writeSamplesJSONL(t)
	if err := runReport([]string{"--in", in, "--out", "-", "--metric", "nope"}); err == nil {
		t.Fatalf("expected error")
	}
}

func TestReport_RejectsUnknownRedactMode(t *testing.T) {
	in := writeSamplesJSONL(t)
	if err := runReport([]string{"--in", in, "--out", "-", "--redact", "nope"}); err == nil {
		t.Fatalf("expected error")
	}
}

func TestReport_AcceptsStaticThreshold(t *testing.T) {
	in := writeSamplesJSONL(t)
	if err := runReport([]string{"--in", in, "--out", "-", "--static-threshold", "cpu=10"}); err != nil {
		t.Fatalf("runReport: %v", err)
	}
}

func TestReport_AcceptsPercentileThreshold(t *testing.T) {
	in := writeSamplesJSONL(t)
	if err := runReport([]string{"--in", in, "--out", "-", "--percentile-threshold", "cpu=95,1.1"}); err != nil {
		t.Fatalf("runReport: %v", err)
	}
}

func TestReport_LastCannotCombineUntil(t *testing.T) {
	in := writeSamplesJSONL(t)
	if err := runReport([]string{"--in", in, "--out", "-", "--last", "1s", "--until", "2026-02-09T00:00:02Z"}); err == nil {
		t.Fatalf("expected error")
	}
}

func TestWatch_RejectsUnknownSeverity(t *testing.T) {
	if err := runWatch([]string{"--duration", "1s", "--min-severity", "nope"}); err == nil {
		t.Fatalf("expected error")
	}
}

func TestWatch_RejectsNegativeDuration(t *testing.T) {
	if err := runWatch([]string{"--duration", "-1s"}); err == nil {
		t.Fatalf("expected error")
	}
}

func TestWatch_RejectsNegativeMaxSamples(t *testing.T) {
	if err := runWatch([]string{"--duration", "1s", "--max-samples", "-1"}); err == nil {
		t.Fatalf("expected error")
	}
}

func TestWatch_RejectsUnknownMetrics(t *testing.T) {
	if err := runWatch([]string{"--duration", "1s", "--metrics", "nope"}); err == nil {
		t.Fatalf("expected error")
	}
}

func TestWatch_RejectsUnknownRedactMode(t *testing.T) {
	if err := runWatch([]string{"--duration", "1s", "--redact", "nope"}); err == nil {
		t.Fatalf("expected error")
	}
}

func TestWatch_RejectsInvalidStaticThreshold(t *testing.T) {
	if err := runWatch([]string{"--duration", "1s", "--static-threshold", "cpu=0"}); err == nil {
		t.Fatalf("expected error")
	}
}

func TestWatch_RejectsNegativeJitter(t *testing.T) {
	if err := runWatch([]string{"--duration", "1s", "--jitter", "-1ms"}); err == nil {
		t.Fatalf("expected error")
	}
}

func TestWatch_RejectsUnknownMetricCooldownMetric(t *testing.T) {
	if err := runWatch([]string{"--duration", "1s", "--metric-cooldown", "nope=1s"}); err == nil {
		t.Fatalf("expected error")
	}
}

func TestWatch_RejectsInvalidPercentileThreshold(t *testing.T) {
	if err := runWatch([]string{"--duration", "1s", "--percentile-threshold", "cpu=95"}); err == nil {
		t.Fatalf("expected error")
	}
}

func TestSelftest_RejectsUnknownFormat(t *testing.T) {
	if err := runSelftest([]string{"--format", "nope"}); err == nil {
		t.Fatalf("expected error")
	}
}

func TestSelftest_RejectsUnknownMetrics(t *testing.T) {
	if err := runSelftest([]string{"--metrics", "nope"}); err == nil {
		t.Fatalf("expected error")
	}
}

func TestCollect_RejectsNegativeMaxSamples(t *testing.T) {
	if err := runCollect([]string{"--once", "--max-samples", "-1"}); err == nil {
		t.Fatalf("expected error")
	}
}

func TestKVLabelsFlag_ParseAndMerge(t *testing.T) {
	var f kvLabelsFlag
	if err := f.Set("env=test"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if err := f.Set("service=api"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if f.m["env"] != "test" || f.m["service"] != "api" {
		t.Fatalf("unexpected labels: %+v", f.m)
	}

	merged := mergeLabels(map[string]string{"region": "us-east-1"}, f.m)
	if merged["region"] != "us-east-1" || merged["env"] != "test" || merged["service"] != "api" {
		t.Fatalf("unexpected merged labels: %+v", merged)
	}
}

func TestMetricCooldownsFlag_ParseAndMerge(t *testing.T) {
	var f metricCooldownsFlag
	if err := f.Set("cpu=5s"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if err := f.Set("disk_read=10s"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if got := f.m["cpu_percent"]; got != 5*time.Second {
		t.Fatalf("unexpected cpu cooldown: %+v", f.m)
	}
	if got := f.m["disk_read_bytes_per_sec"]; got != 10*time.Second {
		t.Fatalf("unexpected disk_read cooldown: %+v", f.m)
	}

	merged := mergeCooldownOverrides(map[string]time.Duration{"mem_used_percent": 3 * time.Second}, f.m)
	if merged["mem_used_percent"] != 3*time.Second || merged["cpu_percent"] != 5*time.Second {
		t.Fatalf("unexpected merged cooldowns: %+v", merged)
	}
}

func TestPercentileThresholdsFlag_ParseAndMerge(t *testing.T) {
	var f percentileThresholdsFlag
	if err := f.Set("cpu=95,1.2"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if err := f.Set("disk_read=99,1.5"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if got := f.m["cpu_percent"]; got.Percentile != 95 || got.Multiplier != 1.2 {
		t.Fatalf("unexpected cpu percentile rule: %+v", got)
	}
	merged := mergePercentileRules(map[string]anomaly.PercentileRule{
		"mem_used_percent": {Percentile: 95, Multiplier: 1.1},
	}, f.m)
	if _, ok := merged["cpu_percent"]; !ok {
		t.Fatalf("expected cpu percentile rule in merge: %+v", merged)
	}
	if _, ok := merged["mem_used_percent"]; !ok {
		t.Fatalf("expected mem percentile rule in merge: %+v", merged)
	}
}

func TestResolveStorageMode_Auto(t *testing.T) {
	mode, err := resolveStorageMode("auto", "samples.db")
	if err != nil || mode != "sqlite" {
		t.Fatalf("expected sqlite mode, got mode=%q err=%v", mode, err)
	}
	mode, err = resolveStorageMode("auto", "samples.jsonl")
	if err != nil || mode != "jsonl" {
		t.Fatalf("expected jsonl mode, got mode=%q err=%v", mode, err)
	}
}

func TestBuildSampleWriter_RejectsMaxSamplesWithJSONL(t *testing.T) {
	if _, err := buildSampleWriter(filepath.Join(t.TempDir(), "samples.jsonl"), "jsonl", false, 1); err == nil {
		t.Fatalf("expected error")
	}
}

func TestNextIntervalWithJitter_NoJitter(t *testing.T) {
	base := 2 * time.Second
	if got := nextIntervalWithJitter(base, 0); got != base {
		t.Fatalf("expected %s, got %s", base, got)
	}
}

func TestNextIntervalWithJitter_Range(t *testing.T) {
	base := 2 * time.Second
	jitter := 500 * time.Millisecond
	for i := 0; i < 100; i++ {
		got := nextIntervalWithJitter(base, jitter)
		if got < base || got > base+jitter {
			t.Fatalf("expected duration in [%s, %s], got %s", base, base+jitter, got)
		}
	}
}
