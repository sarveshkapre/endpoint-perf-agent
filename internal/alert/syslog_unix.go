//go:build !windows

package alert

import (
	"context"
	"encoding/json"
	"fmt"
	"log/syslog"
)

type SyslogSink struct {
	w *syslog.Writer
}

func NewSyslogSink(tag string) (*SyslogSink, error) {
	// Use the local syslog daemon. Priority is chosen per-message based on severity.
	w, err := syslog.New(syslog.LOG_USER|syslog.LOG_INFO, tag)
	if err != nil {
		return nil, err
	}
	return &SyslogSink{w: w}, nil
}

func (s *SyslogSink) Emit(_ context.Context, a Alert) error {
	payload, err := json.Marshal(a)
	if err != nil {
		return err
	}
	msg := string(payload)

	switch a.Severity {
	case "critical":
		return s.w.Crit(msg)
	case "high":
		return s.w.Err(msg)
	case "medium":
		return s.w.Warning(msg)
	case "low":
		return s.w.Info(msg)
	default:
		return s.w.Info(fmt.Sprintf("unknown severity %q: %s", a.Severity, msg))
	}
}

func (s *SyslogSink) Close() error {
	if s.w == nil {
		return nil
	}
	return s.w.Close()
}
