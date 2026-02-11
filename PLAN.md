# Endpoint Perf Agent — Plan

Local-first endpoint performance agent that samples host metrics, detects anomalies against a rolling baseline, and produces explainable reports.

## Features (today)
- Cross-platform CPU/memory/disk/network sampling (gopsutil)
- Per-sample top CPU/memory process attribution for triage context
- Optional labels (`labels`, `--label k=v`) for multi-host and multi-service ingestion
- JSONL storage for easy ingestion
- Rolling z-score anomaly detection with severity + explanations
- Configurable static-threshold alert rules (`static_thresholds` / `--static-threshold`)
- CLI: `collect`, `watch` (alerts to stdout NDJSON or syslog), `analyze` (text/JSON/NDJSON), `report` (Markdown/stdout)

## Top risks / unknowns
- Baselines vary per host; short windows can create false positives/negatives.
- Disk/network counters can reset (reboot, interface changes), skewing rate metrics.
- Platform-specific metric quirks (permissions, containerization, virtualization).

## Commands
See `docs/PROJECT.md` for the canonical command list.

Quick check:
- `make check`

Try it:
- `make build`
- `./bin/epagent collect --duration 30s`
- `./bin/epagent analyze --format json`
- `./bin/epagent report --out -`

## Shipped
- 2026-02-09: Per-sample top-process attribution and anomaly context (timestamp + process details) in analyze/report outputs.
- 2026-02-09: Analysis/report input hardening (`--window`/`--threshold`/`--top`) and line-numbered JSONL parse errors.
- 2026-02-01: `analyze --format json`, `report --out -`, more robust sample handling + clearer explanations.
- 2026-02-01: `--min-severity` + `--top` filters for `analyze`/`report`.
- 2026-02-01: Baseline summaries (mean/stddev/min/max) in Markdown and JSON outputs.

## Next to ship (tight scope)
- Optional SQLite storage with simple retention controls.
- Percentile-based alert rules.
- Sampling jitter to reduce synchronized collection across hosts.
- Per-metric cooldown overrides for watch mode.

## Bigger ideas (tracked)
See `docs/ROADMAP.md`.
