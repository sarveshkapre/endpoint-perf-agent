package report

import (
	"testing"
	"time"

	"github.com/sarveshkapre/endpoint-perf-agent/internal/collector"
)

func TestFilterSamplesByTime_InclusiveBounds(t *testing.T) {
	base := time.Date(2026, 2, 9, 0, 0, 0, 0, time.UTC)
	samples := []collector.MetricSample{
		{Timestamp: base.Add(0 * time.Second)},
		{Timestamp: base.Add(1 * time.Second)},
		{Timestamp: base.Add(2 * time.Second)},
		{Timestamp: base.Add(3 * time.Second)},
	}

	out, err := FilterSamplesByTime(samples, base.Add(1*time.Second), base.Add(2*time.Second))
	if err != nil {
		t.Fatalf("FilterSamplesByTime: %v", err)
	}
	if got, want := len(out), 2; got != want {
		t.Fatalf("expected %d samples, got %d", want, got)
	}
	if !out[0].Timestamp.Equal(base.Add(1*time.Second)) || !out[1].Timestamp.Equal(base.Add(2*time.Second)) {
		t.Fatalf("unexpected timestamps: %+v", []time.Time{out[0].Timestamp, out[1].Timestamp})
	}
}

func TestFilterSamplesByTime_Unbounded(t *testing.T) {
	base := time.Date(2026, 2, 9, 0, 0, 0, 0, time.UTC)
	samples := []collector.MetricSample{
		{Timestamp: base},
		{Timestamp: base.Add(1 * time.Second)},
	}

	out, err := FilterSamplesByTime(samples, time.Time{}, time.Time{})
	if err != nil {
		t.Fatalf("FilterSamplesByTime: %v", err)
	}
	if got, want := len(out), 2; got != want {
		t.Fatalf("expected %d samples, got %d", want, got)
	}
}

func TestFilterSamplesByTime_RejectsSinceAfterUntil(t *testing.T) {
	base := time.Date(2026, 2, 9, 0, 0, 0, 0, time.UTC)
	_, err := FilterSamplesByTime(nil, base.Add(2*time.Second), base.Add(1*time.Second))
	if err == nil {
		t.Fatal("expected error")
	}
}
