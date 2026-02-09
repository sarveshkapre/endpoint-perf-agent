//go:build windows

package alert

import (
	"context"
	"errors"
)

type SyslogSink struct{}

func NewSyslogSink(_ string) (*SyslogSink, error) {
	return nil, errors.New("syslog sink is not supported on windows")
}

func (s *SyslogSink) Emit(_ context.Context, _ Alert) error {
	return errors.New("syslog sink is not supported on windows")
}

func (s *SyslogSink) Close() error { return nil }
