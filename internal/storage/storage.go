package storage

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/sarveshkapre/endpoint-perf-agent/internal/collector"
)

type Writer struct {
	closer io.Closer
	writer *bufio.Writer
}

func NewWriter(path string) (*Writer, error) {
	if path == "-" {
		return NewWriterWithWriter(os.Stdout), nil
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}
	return &Writer{closer: file, writer: bufio.NewWriter(file)}, nil
}

func NewWriterWithWriter(w io.Writer) *Writer {
	return &Writer{writer: bufio.NewWriter(w)}
}

func (w *Writer) Write(sample collector.MetricSample) error {
	payload, err := json.Marshal(sample)
	if err != nil {
		return err
	}
	if _, err := w.writer.Write(append(payload, '\n')); err != nil {
		return err
	}
	return w.writer.Flush()
}

func (w *Writer) Close() error {
	if w.writer != nil {
		_ = w.writer.Flush()
	}
	if w.closer != nil {
		return w.closer.Close()
	}
	return nil
}

func ReadSamples(path string) ([]collector.MetricSample, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	samples := make([]collector.MetricSample, 0)
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		var sample collector.MetricSample
		if err := json.Unmarshal(line, &sample); err != nil {
			return nil, fmt.Errorf("invalid jsonl at line %d: %w", lineNo, err)
		}
		samples = append(samples, sample)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return samples, nil
}
