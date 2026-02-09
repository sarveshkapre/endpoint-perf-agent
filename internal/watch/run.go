package watch

import (
	"context"
	"errors"
	"reflect"
	"time"

	"github.com/sarveshkapre/endpoint-perf-agent/internal/alert"
	"github.com/sarveshkapre/endpoint-perf-agent/internal/collector"
)

type Sampler interface {
	Sample(ctx context.Context) (collector.MetricSample, error)
}

type SampleWriter interface {
	Write(sample collector.MetricSample) error
}

type Runner struct {
	Sampler Sampler
	Engine  *Engine
	Sink    alert.Sink

	Interval time.Duration
	Duration time.Duration

	Writer SampleWriter // optional
}

func (r *Runner) Run(ctx context.Context) error {
	if r.Sampler == nil {
		return errors.New("sampler is required")
	}
	if r.Engine == nil {
		return errors.New("engine is required")
	}
	if r.Sink == nil {
		return errors.New("sink is required")
	}
	if r.Interval <= 0 {
		r.Interval = 5 * time.Second
	}
	if r.Duration < 0 {
		r.Duration = 0
	}

	ticker := time.NewTicker(r.Interval)
	defer ticker.Stop()

	deadline := time.Time{}
	if r.Duration > 0 {
		deadline = time.Now().Add(r.Duration)
	}

	for {
		if !deadline.IsZero() && time.Now().After(deadline) {
			return nil
		}

		sample, err := r.Sampler.Sample(ctx)
		if err != nil {
			return err
		}
		if !isNilInterface(r.Writer) {
			if err := r.Writer.Write(sample); err != nil {
				return err
			}
		}

		for _, a := range r.Engine.Observe(sample) {
			if err := r.Sink.Emit(ctx, a); err != nil {
				return err
			}
		}

		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}

func isNilInterface(v any) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return rv.IsNil()
	default:
		return false
	}
}
