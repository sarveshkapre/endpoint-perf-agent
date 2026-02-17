package watch

import (
	"context"
	"testing"
	"time"

	"github.com/sarveshkapre/endpoint-perf-agent/internal/alert"
	"github.com/sarveshkapre/endpoint-perf-agent/internal/collector"
	"github.com/sarveshkapre/endpoint-perf-agent/internal/storage"
)

type cancelingSampler struct {
	cancel context.CancelFunc
}

func (s *cancelingSampler) Sample(ctx context.Context) (collector.MetricSample, error) {
	s.cancel()
	return collector.MetricSample{
		Timestamp:      time.Now().UTC(),
		CPUPercent:     10,
		MemUsedPercent: 20,
	}, nil
}

type noopSink struct{}

func (noopSink) Emit(_ context.Context, _ alert.Alert) error { return nil }
func (noopSink) Close() error                                { return nil }

func TestRunner_IgnoresTypedNilWriter(t *testing.T) {
	engine, err := NewEngine(5, 3.0, nil, "critical", 0, nil)
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	r := &Runner{
		Sampler:  &cancelingSampler{cancel: cancel},
		Engine:   engine,
		Sink:     noopSink{},
		Interval: 1 * time.Second,
		Duration: 1 * time.Second,
	}

	// This reproduces the classic "typed nil in interface" trap. Previously this
	// would pass the `r.Writer != nil` guard and then panic in Write().
	var w *storage.Writer
	r.Writer = w

	if err := r.Run(ctx); err != nil {
		t.Fatalf("Run: %v", err)
	}
}

func TestRunner_RejectsNegativeJitter(t *testing.T) {
	engine, err := NewEngine(5, 3.0, nil, "critical", 0, nil)
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}

	r := &Runner{
		Sampler:  &cancelingSampler{cancel: func() {}},
		Engine:   engine,
		Sink:     noopSink{},
		Interval: time.Second,
		Jitter:   -1 * time.Second,
	}
	if err := r.Run(context.Background()); err == nil {
		t.Fatalf("expected error")
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
