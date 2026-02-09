# Clone Feature Tracker

## Context Sources
- README and docs
- TODO/FIXME markers in code
- Test and build failures
- Gaps found during codebase exploration

## Candidate Features To Do
- [ ] P0 (Selected) Stream mode: add `epagent watch` to continuously sample and emit anomalies as alerts (stdout NDJSON, optional syslog), with per-metric cooldown/dedupe.
  Score: impact high | effort medium | strategic fit high | differentiation medium | risk low | confidence high
- [ ] P0 (Selected) Overhead controls: add config + CLI flags to disable process attribution (top CPU/mem process scan) for process-dense hosts.
  Score: impact high | effort low | strategic fit high | differentiation low | risk low | confidence high
- [ ] P1 (Selected) CLI integration tests: add command-level tests for flag validation + a minimal end-to-end CLI smoke (collect/analyze/report/watch) using temp files.
  Score: impact medium | effort medium | strategic fit high | differentiation low | risk low | confidence medium
- [ ] P1 Add alert output sinks (structured JSON stream and syslog) for offline analysis results (`analyze`/`report`) as well as `watch`.
  Score: impact high | effort medium | strategic fit high | differentiation medium | risk medium | confidence medium
- [ ] P1 Add configurable static and percentile-based alert rules (in addition to z-score), selectable per metric.
  Score: impact medium | effort high | strategic fit medium | differentiation medium | risk medium | confidence medium
- [ ] P2 Optional SQLite storage mode with retention controls and migration path from JSONL.
  Score: impact medium | effort high | strategic fit medium | differentiation low | risk medium | confidence medium
- [ ] P2 Collector toggles: enable/disable specific metric families (cpu/mem/disk/net) to reduce overhead and noise.
  Score: impact medium | effort medium | strategic fit medium | differentiation low | risk low | confidence medium
- [ ] P3 Improve build noise: investigate/suppress toolchain warnings from dependencies during `go test` on Apple Silicon without hiding real errors.
  Score: impact low | effort low | strategic fit low | differentiation none | risk low | confidence low

## Implemented
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
- Market scan (bounded, untrusted external sources): adjacent tools typically provide a streaming/daemon mode and alert sinks beyond file-based analysis, plus knobs to tune overhead (enable/disable collectors, process monitoring optional).
  Sources:
  - Netdata agent health/alerts: https://learn.netdata.cloud/docs/alerting/health-configuration-reference
  - Prometheus node_exporter metrics collector model: https://github.com/prometheus/node_exporter
  - Glances CLI + exporters (local-first, multi-output): https://nicolargo.github.io/glances/

## Notes
- This file is maintained by the autonomous clone loop.
