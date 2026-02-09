# PROJECT_MEMORY

## Decision Log

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
