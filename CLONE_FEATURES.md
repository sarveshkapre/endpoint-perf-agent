# Clone Feature Tracker

## Context Sources
- README and docs
- TODO/FIXME markers in code
- Test and build failures
- Gaps found during codebase exploration

## Candidate Features To Do
- [ ] P2 Add percentile-based alert rules (p95/p99 vs rolling baseline), selectable per metric family.
  Score: impact medium | effort high | strategic fit high | differentiation medium | risk medium | confidence medium
- [ ] P2 Optional SQLite storage mode with retention controls and a migration path from JSONL.
  Score: impact medium | effort high | strategic fit medium | differentiation low | risk medium | confidence medium
- [ ] P2 Add sampling jitter (bounded) to reduce synchronized collection across hosts.
  Score: impact medium | effort medium | strategic fit medium | differentiation low | risk low | confidence medium
- [ ] P2 Add per-metric cooldown overrides in `watch` to reduce alert suppression for high-value signals.
  Score: impact medium | effort medium | strategic fit high | differentiation low | risk low | confidence medium
- [ ] P2 Add startup warm-up controls (`--warmup-samples`) to suppress noisy early-window anomalies.
  Score: impact medium | effort low | strategic fit high | differentiation low | risk low | confidence high
- [ ] P2 Add anomaly report grouping by metric family and severity counts for faster incident triage.
  Score: impact medium | effort low | strategic fit high | differentiation low | risk low | confidence high
- [ ] P3 Add a `redact` subcommand to transform existing JSONL files (omit/hash sensitive fields) for sharing without rerunning analysis.
  Score: impact low | effort medium | strategic fit medium | differentiation low | risk low | confidence medium
- [ ] P3 Improve build noise: investigate/suppress dependency toolchain warnings on Apple Silicon without hiding real errors.
  Score: impact low | effort low | strategic fit low | differentiation none | risk low | confidence medium
- [ ] P3 Add benchmark coverage for process attribution overhead by process-count tiers.
  Score: impact low | effort medium | strategic fit medium | differentiation low | risk low | confidence medium
- [ ] P3 Add deterministic label ordering for all text/markdown output paths.
  Score: impact low | effort low | strategic fit medium | differentiation none | risk low | confidence high

## Implemented
- [x] 2026-02-11: Added configurable static-threshold rules across `watch`, `analyze`, and `report` (config `static_thresholds` + CLI `--static-threshold metric=value`) with normalized metric aliases, strict validation, and output metadata (`rule_type`, `threshold`).
  Evidence: `internal/config/config.go`, `internal/anomaly/anomaly.go`, `internal/watch/engine.go`, `internal/report/report.go`, `internal/alert/alert.go`, `cmd/epagent/main.go`, tests in `internal/config/config_test.go`, `internal/anomaly/anomaly_test.go`, `internal/watch/engine_test.go`, `internal/report/report_test.go`, `cmd/epagent/main_test.go`, verification `make check`, `collect/analyze/report/watch` smoke commands on `tmp/static-*.json*`.
- [x] 2026-02-10: Added `selftest` command to validate host metric availability/permissions and estimate sampling overhead (including optional process attribution overhead probe).
  Evidence: `cmd/epagent/main.go`, `internal/selftest/selftest.go`, `README.md`, `docs/CHANGELOG.md`, `make check`, local smoke `./bin/epagent selftest --runs 1`.
- [x] 2026-02-10: Added `--redact omit|hash` for `analyze`/`report` outputs and `watch` alerts to omit/hash `host_id` and labels for sharing.
  Evidence: `cmd/epagent/main.go`, `internal/redact/redact.go`, `internal/redact/redact_test.go`, `cmd/epagent/main_test.go`, `README.md`, `docs/CHANGELOG.md`, `make check`.
- [x] 2026-02-10: Added `collect --truncate` and `watch --out ... --truncate` to overwrite sample files instead of appending.
  Evidence: `cmd/epagent/main.go`, `internal/storage/storage.go`, `internal/storage/storage_test.go`, `README.md`, `docs/CHANGELOG.md`, `make check`, local smoke `wc -l` overwrite check.
- [x] 2026-02-10: Added config `labels` and `collect`/`watch` `--label k=v` (repeatable) to tag samples; labels propagate into JSONL samples, alerts, and analysis/report outputs (including per-anomaly labels).
  Evidence: `internal/config/config.go`, `cmd/epagent/main.go`, `internal/collector/collector.go`, `internal/watch/engine.go`, `internal/alert/alert.go`, `internal/anomaly/anomaly.go`, `internal/report/report.go`, `make check`.
- [x] 2026-02-10: Added `analyze`/`report` `--metric cpu|mem|disk|net` (repeatable) to filter output by metric family without re-collecting.
  Evidence: `cmd/epagent/main.go`, `internal/report/metric_family_filter.go`, `internal/report/metric_family_filter_test.go`, `README.md`, `docs/CHANGELOG.md`, `make check`.
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
- Labels are best treated as stable dimensions (env/service/role/region); per-anomaly labels are still useful for mixed-host sample files, but stable labels make summaries and routing much simpler.
- Detector normalization must be reflected in user-visible output; otherwise operations teams can misinterpret baselines and thresholds.
- Line-numbered JSONL parse errors materially reduce diagnosis time for corrupted collection files.
- Gap map refresh (2026-02-11):
  - Missing (now shipped): static absolute thresholds in addition to baseline-relative z-score rules.
  - Weak: jitter and alert cooldown tuning are still global (no per-metric tuning).
  - Parity: metric-family enable/disable, labels/tags, rolling baseline explainers, NDJSON/syslog alert sinks.
  - Differentiator opportunity: lightweight offline-first percentile rules with share-safe redaction and report-grade explanations.
- Market scan refresh (2026-02-11, untrusted external sources): operator baselines consistently include static threshold monitors, anomaly/percentile monitors, collection jitter knobs, and tunable collection intervals/tags.
  Sources:
  - Datadog metric threshold monitor docs: https://docs.datadoghq.com/monitors/types/metric/
  - Datadog anomaly monitor docs: https://docs.datadoghq.com/monitors/types/anomaly/
  - Telegraf configuration (`collection_jitter`, intervals, tags): https://docs.influxdata.com/telegraf/v1/configuration/
  - OpenTelemetry Collector config + hostmetrics receiver (`collection_interval`): https://opentelemetry.io/docs/collector/configuration/ and https://pkg.go.dev/go.opentelemetry.io/collector/receiver/hostmetricsreceiver
  - Elastic Metricbeat module config (`period`, tags/fields): https://www.elastic.co/guide/en/beats/metricbeat/current/configuration-metricbeat.html
- Market scan refresh (2026-02-10, untrusted external sources): consistent labeling/tagging keys (env/service/version or equivalent resource attributes) are a baseline expectation for routing and correlation across metrics/logs/traces.
- Market scan (bounded, untrusted external sources): adjacent tools consistently emphasize (1) enabling/disabling collectors/scrapers to tune overhead, (2) tagging/labeling for multi-host and downstream routing, and (3) interval controls and jitter.
  Sources:
  - Netdata collector configuration: https://learn.netdata.cloud/docs/collecting-metrics/collectors-configuration
  - Netdata performance optimization guide: https://learn.netdata.cloud/docs/netdata-agent/configuration/performance-optimization
  - Prometheus node_exporter collector enable/disable flags: https://github.com/prometheus/node_exporter
  - Glances CLI + exporters (local-first, multi-output): https://nicolargo.github.io/glances/
  - OpenTelemetry Collector hostmetrics receiver (collection interval + enabled scrapers): https://pkg.go.dev/go.opentelemetry.io/collector/receiver/hostmetricsreceiver
  - OpenTelemetry Collector configuration overview (receivers enabled via pipelines): https://opentelemetry.io/docs/collector/configuration/
  - OpenTelemetry Resources (custom resource attributes): https://opentelemetry.io/docs/concepts/resources/
  - OpenTelemetry resource semantic conventions: https://opentelemetry.io/docs/specs/semconv/resource/
  - Telegraf configuration (intervals + tags + plugin filtering): https://docs.influxdata.com/telegraf/v1/configuration/
  - Telegraf `--test` (validate config and emit metrics once): https://docs.influxdata.com/telegraf/v1/commands/telegraf/
  - Elastic Metricbeat module config (period + fields + tags): https://www.elastic.co/guide/en/beats/metricbeat/current/configuration-metricbeat.html
  - Datadog unified service tagging (`env`, `service`, `version`): https://docs.datadoghq.com/getting_started/tagging/unified_service_tagging/

## Notes
- This file is maintained by the autonomous clone loop.
