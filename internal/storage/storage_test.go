package storage

import (
	"os"
	"strings"
	"testing"
)

func TestReadSamplesSkipsBlankLines(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "samples-*.jsonl")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	defer f.Close()

	payload := `{"timestamp":"2026-02-01T00:00:00Z","host_id":"","cpu_percent":1,"mem_used_percent":2,"disk_used_percent":3,"disk_read_bytes":4,"disk_write_bytes":5,"net_rx_bytes":6,"net_tx_bytes":7}

{"timestamp":"2026-02-01T00:00:01Z","host_id":"","cpu_percent":8,"mem_used_percent":9,"disk_used_percent":10,"disk_read_bytes":11,"disk_write_bytes":12,"net_rx_bytes":13,"net_tx_bytes":14}
`
	if _, err := f.WriteString(payload); err != nil {
		t.Fatalf("WriteString: %v", err)
	}

	samples, err := ReadSamples(f.Name())
	if err != nil {
		t.Fatalf("ReadSamples: %v", err)
	}
	if got, want := len(samples), 2; got != want {
		t.Fatalf("expected %d samples, got %d", want, got)
	}
}

func TestReadSamplesReportsLineNumberOnParseError(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "samples-*.jsonl")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	defer f.Close()

	payload := `{"timestamp":"2026-02-01T00:00:00Z","host_id":"","cpu_percent":1,"mem_used_percent":2,"disk_used_percent":3,"disk_read_bytes":4,"disk_write_bytes":5,"net_rx_bytes":6,"net_tx_bytes":7}
not-json
`
	if _, err := f.WriteString(payload); err != nil {
		t.Fatalf("WriteString: %v", err)
	}

	_, err = ReadSamples(f.Name())
	if err == nil {
		t.Fatal("expected parse error")
	}
	if !strings.Contains(err.Error(), "line 2") {
		t.Fatalf("expected line number in error, got: %v", err)
	}
}
