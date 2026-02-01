package report

import (
	"testing"

	"github.com/sarveshkapre/endpoint-perf-agent/internal/anomaly"
)

func TestApplyFiltersMinSeverity(t *testing.T) {
	in := AnalysisResult{
		TotalAnomalies: 3,
		Anomalies: []anomaly.Anomaly{
			{Name: "a", Severity: "low", ZScore: 10},
			{Name: "b", Severity: "medium", ZScore: 9},
			{Name: "c", Severity: "high", ZScore: 8},
		},
	}

	out, err := ApplyFilters(in, "high", 0)
	if err != nil {
		t.Fatalf("ApplyFilters: %v", err)
	}
	if got, want := out.TotalAnomalies, 3; got != want {
		t.Fatalf("expected TotalAnomalies=%d, got %d", want, got)
	}
	if got, want := len(out.Anomalies), 1; got != want {
		t.Fatalf("expected %d anomalies, got %d", want, got)
	}
	if out.Anomalies[0].Name != "c" {
		t.Fatalf("expected high severity anomaly to remain")
	}
}

func TestApplyFiltersTopByAbsZScore(t *testing.T) {
	in := AnalysisResult{
		TotalAnomalies: 3,
		Anomalies: []anomaly.Anomaly{
			{Name: "a", Severity: "low", ZScore: -100},
			{Name: "b", Severity: "low", ZScore: 50},
			{Name: "c", Severity: "low", ZScore: -60},
		},
	}

	out, err := ApplyFilters(in, "low", 2)
	if err != nil {
		t.Fatalf("ApplyFilters: %v", err)
	}
	if got, want := len(out.Anomalies), 2; got != want {
		t.Fatalf("expected %d anomalies, got %d", want, got)
	}
	// Must include the two biggest |z|: 100 and 60 (in any order).
	names := map[string]bool{out.Anomalies[0].Name: true, out.Anomalies[1].Name: true}
	if !names["a"] || !names["c"] {
		t.Fatalf("expected anomalies a and c, got: %+v", out.Anomalies)
	}
}

func TestApplyFiltersRejectsUnknownSeverity(t *testing.T) {
	_, err := ApplyFilters(AnalysisResult{}, "bogus", 0)
	if err == nil {
		t.Fatal("expected error")
	}
}
