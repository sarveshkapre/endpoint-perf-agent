package alert

import (
	"context"
	"encoding/json"
	"io"
	"sync"
	"time"

	"github.com/sarveshkapre/endpoint-perf-agent/internal/anomaly"
)

type Alert struct {
	Timestamp     time.Time                   `json:"timestamp"`
	HostID        string                      `json:"host_id,omitempty"`
	Metric        string                      `json:"metric"`
	Value         float64                     `json:"value"`
	Mean          float64                     `json:"mean"`
	Stddev        float64                     `json:"stddev"`
	ZScore        float64                     `json:"zscore"`
	Severity      string                      `json:"severity"`
	Explanation   string                      `json:"explanation"`
	TopCPUProcess *anomaly.ProcessAttribution `json:"top_cpu_process,omitempty"`
	TopMemProcess *anomaly.ProcessAttribution `json:"top_mem_process,omitempty"`
}

type Sink interface {
	Emit(ctx context.Context, a Alert) error
	Close() error
}

type StdoutSink struct {
	mu  sync.Mutex
	enc *json.Encoder
	w   io.Writer
}

func NewStdoutSink(w io.Writer) *StdoutSink {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	return &StdoutSink{enc: enc, w: w}
}

func (s *StdoutSink) Emit(_ context.Context, a Alert) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.enc.Encode(a)
}

func (s *StdoutSink) Close() error { return nil }

func SeverityRank(severity string) (int, bool) {
	switch severity {
	case "low":
		return 1, true
	case "medium":
		return 2, true
	case "high":
		return 3, true
	case "critical":
		return 4, true
	default:
		return 0, false
	}
}
