# PROJECT_MEMORY

## Decision Log

### 2026-02-09 - Add stdout sample output and NDJSON anomaly emission for offline pipelines
- Decision:
  - Add `collect --out -` to stream JSONL samples to stdout.
  - Add `analyze --format ndjson` with `--sink stdout|syslog` to emit one alert per anomaly as JSON for easy piping into log pipelines.
- Why: CLI-first operators often want to pipe data without managing files; offline anomaly output needs a streaming-friendly format for downstream tooling.
- Evidence:
  - Code: `internal/storage/storage.go`, `cmd/epagent/main.go`
  - Tests: `internal/storage/storage_test.go`, `cmd/epagent/main_test.go`
  - Local smoke:
    - `./bin/epagent collect --once --out - --process-attribution=false --metrics cpu,mem | head -n 1`
    - `./bin/epagent analyze --in tmp/anom.jsonl --window 5 --threshold 2.5 --format ndjson --sink stdout`
    - `./bin/epagent analyze --in tmp/anom.jsonl --window 5 --threshold 2.5 --format ndjson --sink syslog --syslog-tag epagent-test`
- Commit: `8b1f21d`
- Confidence: high
- Trust label: verified-local

### 2026-02-09 - Add metric family allow-listing and persist per-sample collection scope
- Decision: Add `enabled_metrics` (config) and `--metrics` (collect/watch) to allow-list metric families (cpu/mem/disk/net), persist `metric_families` in each sample, and make `analyze`/`report` and `watch` respect what was collected.
- Why: Production hosts vary; being able to disable noisy or expensive collectors improves usability and reduces overhead while keeping analysis correct for partial datasets.
- Evidence:
  - Code: `internal/config/config.go`, `internal/collector/collector.go`, `internal/report/report.go`, `internal/watch/engine.go`, `cmd/epagent/main.go`
  - Tests: `internal/config/config_test.go`, `internal/report/report_test.go`, `internal/watch/engine_test.go`
  - Local smoke:
    - `./bin/epagent collect --once --out tmp/smoke-metrics.XXXXXX.jsonl --process-attribution=false --metrics cpu,mem`
    - `./bin/epagent analyze --in tmp/smoke-metrics.XXXXXX.jsonl --format json --window 5 --threshold 3`
    - `./bin/epagent report --in tmp/smoke-metrics.XXXXXX.jsonl --out tmp/report.md --min-severity low --window 5 --threshold 3`
- Commit: `e5235ea`, `8f4837e`
- Confidence: high
- Trust label: verified-local

### 2026-02-09 - Fix `watch` optional sample writer typed-nil panic
- Decision: Ensure `watch` does not panic when `--out` is unset by (1) passing a nil interface from the CLI and (2) hardening `Runner` to treat typed-nil interface values as nil.
- Why: `watch` should be safe by default; optional sample output must never crash the agent.
- Evidence:
  - Code: `cmd/epagent/main.go`, `internal/watch/run.go`
  - Tests: `internal/watch/run_test.go`
  - Local smoke: `./bin/epagent watch --duration 2s --interval 1s --metrics cpu,mem --process-attribution=false --sink stdout --min-severity critical > tmp/watch-metrics.ndjson`
- Commit: `e5235ea`, `8f4837e`
- Confidence: high
- Trust label: verified-local

### 2026-02-09 - Add process attribution context to anomaly outputs
- Decision: Capture top CPU and top memory process during collection and propagate this context to anomaly records in `analyze` and `report` output.
- Why: Metric-only anomalies slow triage. Process context lets operators quickly identify likely offenders without extra tooling.
- Evidence:
  - Code: `internal/collector/collector.go`, `internal/anomaly/anomaly.go`, `internal/report/report.go`
  - Tests: `internal/report/report_test.go`
  - Local smoke: `./bin/epagent collect --once --out tmp/smoke-metrics.jsonl` then `./bin/epagent analyze --in tmp/smoke-metrics.jsonl --format json`
- Commit: `cfc8a21`
- Confidence: high
- Trust label: verified-local
- Follow-ups:
  - Add optional process allow/deny filters for noisy hosts.
  - Measure collection overhead on hosts with very high process counts.

### 2026-02-09 - Harden analysis parameter behavior and JSONL diagnostics
- Decision: Normalize detector parameters to safe defaults and add explicit validation for negative CLI values; include JSONL line numbers in parse failures.
- Why: Prevent confusing output mismatch and reduce time to diagnose malformed sample files.
- Evidence:
  - Code: `cmd/epagent/main.go`, `internal/report/report.go`, `internal/report/filters.go`, `internal/storage/storage.go`
  - Tests: `internal/report/filters_test.go`, `internal/storage/storage_test.go`, `internal/report/report_test.go`
  - Validation: `./bin/epagent analyze --in tmp/smoke-metrics.jsonl --top -1` returns validation error.
- Commit: `cfc8a21`
- Confidence: high
- Trust label: verified-local
- Follow-ups:
  - Add command-level tests for CLI flag validation paths.

### 2026-02-09 - Add repository AGENTS contract and keep tracker current
- Decision: Track the autonomous operating contract in-repo as `AGENTS.md`, and keep the clone task tracker current (including bounded market scan notes with links).
- Why: The maintainer loop needs a stable, auditable contract and an always-current backlog that stays aligned with shipped behavior.
- Evidence:
  - Docs: `AGENTS.md`, `CLONE_FEATURES.md`
- Commit: `e7e40f0`
- Confidence: high
- Trust label: verified-local

### 2026-02-09 - Add process attribution overhead control
- Decision: Add config `process_attribution` (default true) and `collect --process-attribution=false` to disable per-sample process scans on process-dense hosts.
- Why: Scanning all processes each interval can be expensive; production use needs a knob to trade triage context for overhead.
- Evidence:
  - Code: `internal/config/config.go`, `internal/collector/collector.go`, `cmd/epagent/main.go`
  - Docs: `README.md`, `docs/CHANGELOG.md`
- Commit: `4e4b618`
- Confidence: high
- Trust label: verified-local
 - Follow-up shipped (same day): updated CLI parsing so config defaults are respected unless explicitly overridden by flags (fixes bool-flag default overriding config).
   Commit: `2b0b2f4`

### 2026-02-09 - Add streaming watch mode + alert sinks
- Decision: Add `epagent watch` to continuously sample and emit anomaly alerts to stdout (NDJSON) or syslog, with per-metric cooldown and optional JSONL sample output.
- Why: File-based collection + offline analysis is useful, but streaming alerts materially improves time-to-detection and fits common ops workflows (pipe/ship, syslog).
- Evidence:
  - Code: `cmd/epagent/main.go`, `internal/watch/engine.go`, `internal/watch/run.go`, `internal/alert/alert.go`
  - Tests: `cmd/epagent/main_test.go`, `internal/watch/engine_test.go`
  - Local smoke: `./bin/epagent watch --duration 5s --interval 1s --sink stdout --min-severity medium --process-attribution=false --out tmp/watch-metrics.jsonl`
- Commit: `6f808b0`
- Confidence: high
- Trust label: verified-local

## Verification Evidence
- `make check` (pass)
- `gh run list -L 5 --branch main` (pass; latest `ci`, `secret-scan`, `codeql` runs succeeded for `main` push)
- `./bin/epagent collect --once --out - --process-attribution=false --metrics cpu,mem | head -n 1` (pass)
- `./bin/epagent analyze --in tmp/anom.jsonl --window 5 --threshold 2.5 --format ndjson --sink stdout` (pass; emitted 1 alert in that run)
- `./bin/epagent analyze --in tmp/anom.jsonl --window 5 --threshold 2.5 --format ndjson --sink syslog --syslog-tag epagent-test` (pass)
- `./bin/epagent collect --once --out tmp/smoke-metrics.XXXXXX.jsonl --process-attribution=false --metrics cpu,mem` (pass)
- `./bin/epagent analyze --in tmp/smoke-metrics.XXXXXX.jsonl --format json --window 5 --threshold 3` (pass)
- `./bin/epagent report --in tmp/smoke-metrics.XXXXXX.jsonl --out tmp/report.md --min-severity low --window 5 --threshold 3` (pass)
- `./bin/epagent watch --duration 2s --interval 1s --metrics cpu,mem --process-attribution=false --sink stdout --min-severity critical > tmp/watch-metrics.ndjson` (pass; 0 alerts emitted in that run)
- `./bin/epagent collect --once --out tmp/smoke-metrics.jsonl --process-attribution=false` (pass)
- `./bin/epagent analyze --in tmp/smoke-metrics.jsonl --format json --window 5 --threshold 3 --min-severity low` (pass)
- `./bin/epagent report --in tmp/smoke-metrics.jsonl --out tmp/report.md --min-severity low` (pass)
- `./bin/epagent watch --duration 5s --interval 1s --min-severity medium --sink stdout --process-attribution=false --out tmp/watch-metrics.jsonl` (pass; 0 alerts emitted in that run)
- `gh run list -L 5` (pass; latest `ci`, `secret-scan`, `codeql` runs succeeded for `main` push)
- `./bin/epagent collect --once --out tmp/smoke2.jsonl --process-attribution=false` (pass)
- `./bin/epagent watch --duration 2s --interval 1s --process-attribution=false --out tmp/watch2.jsonl --sink stdout --min-severity critical` (pass; 0 alerts emitted in that run)
