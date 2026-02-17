package storage

import (
	"bytes"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sarveshkapre/endpoint-perf-agent/internal/collector"
	_ "modernc.org/sqlite"
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

func TestWriterWithWriterWritesJSONL(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriterWithWriter(&buf)

	sample := collector.MetricSample{
		Timestamp:      time.Date(2026, 2, 9, 0, 0, 0, 0, time.UTC),
		HostID:         "test",
		CPUPercent:     1,
		MemUsedPercent: 2,
	}
	if err := w.Write(sample); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if !strings.HasSuffix(buf.String(), "\n") {
		t.Fatalf("expected newline-terminated JSONL")
	}
	if !strings.Contains(buf.String(), `"cpu_percent":1`) {
		t.Fatalf("expected payload to include cpu_percent, got: %s", buf.String())
	}
}

func TestNewWriterWithOptions_TruncateOverwrites(t *testing.T) {
	path := t.TempDir() + "/samples.jsonl"

	w1, err := NewWriterWithOptions(path, true)
	if err != nil {
		t.Fatalf("NewWriterWithOptions(append): %v", err)
	}
	if err := w1.Write(collector.MetricSample{Timestamp: time.Now().UTC(), HostID: "a", CPUPercent: 1}); err != nil {
		t.Fatalf("Write: %v", err)
	}
	_ = w1.Close()

	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if len(before) == 0 {
		t.Fatalf("expected non-empty file after first write")
	}

	w2, err := NewWriterWithOptions(path, false)
	if err != nil {
		t.Fatalf("NewWriterWithOptions(truncate): %v", err)
	}
	if err := w2.Write(collector.MetricSample{Timestamp: time.Now().UTC(), HostID: "b", CPUPercent: 2}); err != nil {
		t.Fatalf("Write: %v", err)
	}
	_ = w2.Close()

	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if bytes.Contains(after, []byte(`"host_id":"a"`)) {
		t.Fatalf("expected truncate writer to overwrite old contents, got: %s", string(after))
	}
	if !bytes.Contains(after, []byte(`"host_id":"b"`)) {
		t.Fatalf("expected new contents, got: %s", string(after))
	}
}

func TestSQLiteWriter_RetentionMaxSamples(t *testing.T) {
	path := filepath.Join(t.TempDir(), "samples.db")
	w, err := NewSQLiteWriter(path, 2, false)
	if err != nil {
		t.Fatalf("NewSQLiteWriter: %v", err)
	}

	base := time.Date(2026, 2, 10, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 3; i++ {
		if err := w.Write(collector.MetricSample{
			Timestamp:      base.Add(time.Duration(i) * time.Second),
			HostID:         string(rune('a' + i)),
			CPUPercent:     float64(10 + i),
			MemUsedPercent: 20,
		}); err != nil {
			t.Fatalf("Write: %v", err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM samples`).Scan(&count); err != nil {
		t.Fatalf("QueryRow count: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected retained row count 2, got %d", count)
	}
}

func TestSQLiteWriter_Truncate(t *testing.T) {
	path := filepath.Join(t.TempDir(), "samples.db")

	w1, err := NewSQLiteWriter(path, 0, false)
	if err != nil {
		t.Fatalf("NewSQLiteWriter: %v", err)
	}
	if err := w1.Write(collector.MetricSample{Timestamp: time.Now().UTC(), HostID: "first", CPUPercent: 1}); err != nil {
		t.Fatalf("Write: %v", err)
	}
	_ = w1.Close()

	w2, err := NewSQLiteWriter(path, 0, true)
	if err != nil {
		t.Fatalf("NewSQLiteWriter truncate: %v", err)
	}
	if err := w2.Write(collector.MetricSample{Timestamp: time.Now().UTC(), HostID: "second", CPUPercent: 2}); err != nil {
		t.Fatalf("Write: %v", err)
	}
	_ = w2.Close()

	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	defer db.Close()

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM samples`).Scan(&count); err != nil {
		t.Fatalf("QueryRow count: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected row count 1 after truncate, got %d", count)
	}
}

func TestSQLiteWriter_RejectsNegativeMaxSamples(t *testing.T) {
	if _, err := NewSQLiteWriter(filepath.Join(t.TempDir(), "samples.db"), -1, false); err == nil {
		t.Fatalf("expected error")
	}
}
