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
- [x] Markdown report
- [x] Unit tests for detector
- [x] CI with lint/test/build + security scans

## Risks
- Metric baselines vary per host; false positives with short windows.
- Disk/network deltas may reset after reboot.

## Next Milestones
- Per-process attribution for top offenders
- Optional SQLite storage
- Configurable alert rules (percentile, static thresholds)
