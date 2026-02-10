# PLAN

## Goal
Ship a local-first endpoint performance agent that samples host metrics, detects anomalies against a rolling baseline, and produces explainable reports.

## Stack
- Go 1.22
- gopsutil v3 for cross-platform metrics
- JSONL for storage

## Architecture
- `cmd/epagent`: CLI entrypoint
- `internal/collector`: sampling logic
- `internal/anomaly`: rolling z-score detector
- `internal/storage`: JSONL persistence
- `internal/report`: analysis and markdown report output

## MVP Checklist
- [x] Sample CPU, memory, disk, network metrics
- [x] JSONL storage
- [x] Rolling anomaly detection + severity
- [x] CLI: collect/analyze/report
- [x] Analyze output format: JSON (`analyze --format json`)
- [x] Markdown report
- [x] Report to stdout (`report --out -`)
- [x] Sort samples by timestamp; tolerate blank lines in JSONL
- [x] Direction-aware anomaly explanations
- [x] Filter + limit anomalies (`--min-severity`, `--top`)
- [x] Unit tests for detector
- [x] CI with lint/test/build + security scans

## Risks
- Metric baselines vary per host; false positives with short windows.
- Disk/network deltas may reset after reboot.

## Next Milestones
- Optional SQLite storage
- Configurable alert rules (percentile, static thresholds)
- `selftest` / overhead benchmarking to validate host readiness and collection cost
- Optional redaction mode for sharing outputs (omit/hash host_id + labels)

## Shipped
- 2026-02-10: Added config `labels` and `collect`/`watch` `--label k=v` (repeatable); labels propagate into alerts and reports.
- 2026-02-10: Added `analyze`/`report` `--metric cpu|mem|disk|net` (repeatable) to filter output by metric family.
- 2026-02-09: Per-sample top-process attribution and anomaly context (timestamp + process details) in analyze/report outputs.
- 2026-02-09: Analysis/report input hardening (`--window`/`--threshold`/`--top`) and line-numbered JSONL parse errors.
- 2026-02-01: `analyze --format json`, `report --out -`, more robust sample handling + clearer explanations.
- 2026-02-01: `--min-severity` + `--top` filters for `analyze`/`report`.
- 2026-02-01: Baseline summaries (mean/stddev/min/max) in Markdown and JSON outputs.
