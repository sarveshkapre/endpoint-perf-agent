package report

import (
	"fmt"
	"time"

	"github.com/sarveshkapre/endpoint-perf-agent/internal/collector"
)

// FilterSamplesByTime returns only samples in the inclusive [since, until] time range.
// Zero values for since/until mean "unbounded".
func FilterSamplesByTime(samples []collector.MetricSample, since, until time.Time) ([]collector.MetricSample, error) {
	if !since.IsZero() && !until.IsZero() && since.After(until) {
		return nil, fmt.Errorf("since must be less than or equal to until")
	}
	if since.IsZero() && until.IsZero() {
		return samples, nil
	}

	out := make([]collector.MetricSample, 0, len(samples))
	for _, s := range samples {
		ts := s.Timestamp
		if !since.IsZero() && ts.Before(since) {
			continue
		}
		if !until.IsZero() && ts.After(until) {
			continue
		}
		out = append(out, s)
	}
	return out, nil
}
