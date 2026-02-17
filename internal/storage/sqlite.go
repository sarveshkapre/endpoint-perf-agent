package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/sarveshkapre/endpoint-perf-agent/internal/collector"
	_ "modernc.org/sqlite"
)

type SQLiteWriter struct {
	db         *sql.DB
	maxSamples int
}

func NewSQLiteWriter(path string, maxSamples int, truncate bool) (*SQLiteWriter, error) {
	if maxSamples < 0 {
		return nil, fmt.Errorf("max samples must be greater than or equal to zero")
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(`PRAGMA journal_mode=WAL;`); err != nil {
		_ = db.Close()
		return nil, err
	}
	if _, err := db.Exec(`
CREATE TABLE IF NOT EXISTS samples (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	timestamp TEXT NOT NULL,
	payload TEXT NOT NULL
)`); err != nil {
		_ = db.Close()
		return nil, err
	}
	if truncate {
		if _, err := db.Exec(`DELETE FROM samples`); err != nil {
			_ = db.Close()
			return nil, err
		}
	}
	return &SQLiteWriter{db: db, maxSamples: maxSamples}, nil
}

func (w *SQLiteWriter) Write(sample collector.MetricSample) error {
	payload, err := json.Marshal(sample)
	if err != nil {
		return err
	}
	tx, err := w.db.Begin()
	if err != nil {
		return err
	}
	if _, err := tx.Exec(`INSERT INTO samples (timestamp, payload) VALUES (?, ?)`, sample.Timestamp.Format("2006-01-02T15:04:05.999999999Z07:00"), string(payload)); err != nil {
		_ = tx.Rollback()
		return err
	}
	if w.maxSamples > 0 {
		if _, err := tx.Exec(`
DELETE FROM samples
WHERE id IN (
	SELECT id FROM samples
	ORDER BY id DESC
	LIMIT -1 OFFSET ?
)`, w.maxSamples); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func (w *SQLiteWriter) Close() error {
	if w == nil || w.db == nil {
		return nil
	}
	return w.db.Close()
}

func ReadSamplesFromSQLite(path string) ([]collector.MetricSample, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query(`SELECT payload FROM samples ORDER BY id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	samples := make([]collector.MetricSample, 0)
	rowNum := 0
	for rows.Next() {
		rowNum++
		var payload string
		if err := rows.Scan(&payload); err != nil {
			return nil, err
		}
		var sample collector.MetricSample
		if err := json.Unmarshal([]byte(payload), &sample); err != nil {
			return nil, fmt.Errorf("invalid sqlite sample payload at row %d: %w", rowNum, err)
		}
		samples = append(samples, sample)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return samples, nil
}
