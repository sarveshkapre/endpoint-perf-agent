package storage

import (
	"bufio"
	"encoding/json"
	"os"

	"github.com/sarveshkapre/endpoint-perf-agent/internal/collector"
)

type Writer struct {
	file   *os.File
	writer *bufio.Writer
}

func NewWriter(path string) (*Writer, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}
	return &Writer{file: file, writer: bufio.NewWriter(file)}, nil
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
	if w.file != nil {
		return w.file.Close()
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
	samples := make([]collector.MetricSample, 0)
	for scanner.Scan() {
		var sample collector.MetricSample
		if err := json.Unmarshal(scanner.Bytes(), &sample); err != nil {
			return nil, err
		}
		samples = append(samples, sample)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return samples, nil
}
