# Clone Feature Tracker

## Context Sources
- README and docs
- TODO/FIXME markers in code
- Test and build failures
- Gaps found during codebase exploration

## Candidate Features To Do
- [ ] P1 (Selected Next) Add optional key/value labels (`--label k=v`, config labels) for multi-host/multi-service ingestion; propagate labels into alerts and reports.
  Score: impact high | effort medium | strategic fit high | differentiation low | risk medium | confidence medium
- [ ] P2 Add configurable static-threshold alert rules (in addition to z-score), selectable per metric.
  Score: impact medium | effort medium | strategic fit medium | differentiation medium | risk medium | confidence medium
- [ ] P2 Add percentile-based alert rules (p95/p99 vs rolling baseline) selectable per metric.
  Score: impact medium | effort high | strategic fit medium | differentiation medium | risk medium | confidence medium
- [ ] P2 Optional SQLite storage mode with retention controls and a migration path from JSONL.
  Score: impact medium | effort high | strategic fit medium | differentiation low | risk medium | confidence medium
- [ ] P3 Add `--metric` include filter for `analyze`/`report` to focus on a subset of metrics (e.g. only cpu + net).
  Score: impact low | effort low | strategic fit medium | differentiation low | risk low | confidence medium
- [ ] P3 Add a lightweight `bench`/`selftest` command to estimate collection overhead (interval jitter, process attribution cost) and validate metrics availability on the host.
  Score: impact low | effort medium | strategic fit medium | differentiation low | risk low | confidence low
- [ ] P3 Improve build noise: investigate/suppress toolchain warnings from dependencies during `go test` on Apple Silicon without hiding real errors.
  Score: impact low | effort low | strategic fit low | differentiation none | risk low | confidence low

## Implemented
- [x] 2026-02-09: Added `collect`/`watch` `--host-id` override to simplify multi-host ingestion and ad-hoc runs without config edits.
  Evidence: `cmd/epagent/main.go`, `README.md`, `docs/CHANGELOG.md`, `make check`.
- [x] 2026-02-09: Added `analyze`/`report` time window filtering via `--since`/`--until` (RFC3339) and `--last <duration>` convenience filtering.
  Evidence: `cmd/epagent/main.go`, `internal/report/sample_filter.go`, `cmd/epagent/main_test.go`, `internal/report/sample_filter_test.go`, `README.md`, `docs/CHANGELOG.md`, `make check`.
- [x] 2026-02-09: Added `AGENTS.md` operating contract to the repository root and updated the clone tracker with bounded market scan notes.
  Evidence: `AGENTS.md`, `CLONE_FEATURES.md`.
- [x] 2026-02-09: Added `collect --process-attribution=false` and config `process_attribution` to disable per-sample process scans when overhead is a concern.
  Evidence: `internal/config/config.go`, `internal/collector/collector.go`, `cmd/epagent/main.go`, `README.md`, `docs/CHANGELOG.md`.
- [x] 2026-02-09: Added metric family allow-listing (`enabled_metrics` / `--metrics`) and persisted per-sample metric family metadata so `analyze`/`report` can respect collection scope.
  Evidence: `internal/config/config.go`, `internal/collector/collector.go`, `internal/report/report.go`, `internal/watch/engine.go`, `cmd/epagent/main.go`, `README.md`, `docs/CHANGELOG.md`.
- [x] 2026-02-09: Added `collect --out -` to stream JSONL samples to stdout.
  Evidence: `internal/storage/storage.go`, `internal/storage/storage_test.go`, `cmd/epagent/main.go`, `README.md`, `docs/CHANGELOG.md`.
- [x] 2026-02-09: Added `analyze --format ndjson` plus `--sink stdout|syslog` to emit one JSON alert per anomaly.
  Evidence: `cmd/epagent/main.go`, `cmd/epagent/main_test.go`, `README.md`, `docs/CHANGELOG.md`.
- [x] 2026-02-09: Added `watch` command to stream anomaly alerts to stdout (NDJSON) or syslog, with per-metric cooldown and optional JSONL sample output.
  Evidence: `cmd/epagent/main.go`, `internal/watch/engine.go`, `internal/watch/run.go`, `internal/alert/alert.go`, `internal/alert/syslog_unix.go`, `internal/alert/syslog_windows.go`.
- [x] 2026-02-09: Fixed `watch` panic when `--out` is unset (typed-nil writer assigned to interface).
  Evidence: `cmd/epagent/main.go`, `internal/watch/run.go`, `internal/watch/run_test.go`.
- [x] 2026-02-09: Added command-level CLI tests for validation and basic flows (analyze/report/watch) without requiring host metrics.
  Evidence: `cmd/epagent/main_test.go`, `internal/watch/engine_test.go`.
- [x] 2026-02-09: Added per-sample process attribution (`top_cpu_process`, `top_mem_process`) in collector output.
  Evidence: `internal/collector/collector.go`, smoke JSONL sample from `epagent collect --once`.
- [x] 2026-02-09: Added anomaly context enrichment (timestamp + top CPU/memory process) in analysis/report output paths.
  Evidence: `internal/anomaly/anomaly.go`, `internal/report/report.go`, `internal/report/report_test.go`.
- [x] 2026-02-09: Hardened analysis flag handling by normalizing detector params and validating negative CLI values.
  Evidence: `cmd/epagent/main.go`, `internal/report/report.go`, `internal/report/filters.go`, `internal/report/filters_test.go`.
- [x] 2026-02-09: Improved JSONL diagnostics with line-numbered parse errors.
  Evidence: `internal/storage/storage.go`, `internal/storage/storage_test.go`.
- [x] 2026-02-09: Added project memory and incident tracking docs.
  Evidence: `PROJECT_MEMORY.md`, `INCIDENTS.md`.
- [x] 2026-02-09: Synced core product docs with shipped behavior.
  Evidence: `README.md`, `docs/CHANGELOG.md`, `docs/ROADMAP.md`, `docs/PROJECT.md`, `docs/PLAN.md`, `PLAN.md`, `UPDATE.md`.

## Insights
- Process attribution substantially improves first-pass triage, but can bias toward transient short-lived processes; follow-up benchmarking is needed on process-dense hosts.
- Detector normalization must be reflected in user-visible output; otherwise operations teams can misinterpret baselines and thresholds.
- Line-numbered JSONL parse errors materially reduce diagnosis time for corrupted collection files.
- Market scan (bounded, untrusted external sources): adjacent tools consistently emphasize (1) enabling/disabling collectors/scrapers to tune overhead, (2) tagging/labeling for multi-host and downstream routing, and (3) interval controls and jitter.
  Sources:
  - Netdata collector configuration: https://learn.netdata.cloud/docs/collecting-metrics/collectors-configuration
  - Netdata performance optimization guide: https://learn.netdata.cloud/docs/netdata-agent/configuration/performance-optimization
  - Prometheus node_exporter collector enable/disable flags: https://github.com/prometheus/node_exporter
  - Glances CLI + exporters (local-first, multi-output): https://nicolargo.github.io/glances/
  - OpenTelemetry Collector hostmetrics receiver (collection interval + enabled scrapers): https://pkg.go.dev/go.opentelemetry.io/collector/receiver/hostmetricsreceiver
  - OpenTelemetry Collector configuration overview (receivers enabled via pipelines): https://opentelemetry.io/docs/collector/configuration/
  - Telegraf configuration (intervals + tags + plugin filtering): https://docs.influxdata.com/telegraf/v1/configuration/
  - Elastic Metricbeat module config (period + fields + tags): https://www.elastic.co/guide/en/beats/metricbeat/current/configuration-metricbeat.html

## Notes
- This file is maintained by the autonomous clone loop.
