package main

import (
	"os"
	"path/filepath"
	"testing"
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
