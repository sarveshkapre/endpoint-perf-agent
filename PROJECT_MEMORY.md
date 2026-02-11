# PROJECT_MEMORY

## Decision Log

### 2026-02-11 - Add static-threshold rules with rule metadata and CLI/config support
- Decision:
  - Added static-threshold rules for `watch`, `analyze`, and `report` via config `static_thresholds` and CLI `--static-threshold metric=value` (repeatable).
  - Added `rule_type` and `threshold` metadata to analysis anomalies and emitted alerts so downstream tooling can distinguish z-score and static-threshold triggers.
  - Updated report/summary formatting to explain static-threshold findings explicitly and refreshed planning docs to keep backlog realistic.
- Why: Rolling z-score alone can miss absolute saturation conditions when baselines drift; production operators need absolute guardrails for known ceilings (CPU%, memory%, disk%, throughput).
- Evidence:
  - Code: `cmd/epagent/main.go`, `internal/config/config.go`, `internal/anomaly/anomaly.go`, `internal/watch/engine.go`, `internal/report/report.go`, `internal/alert/alert.go`
  - Tests: `cmd/epagent/main_test.go`, `internal/config/config_test.go`, `internal/anomaly/anomaly_test.go`, `internal/watch/engine_test.go`, `internal/report/report_test.go`
  - Local smoke:
    - `./bin/epagent collect --duration 6s --interval 1s --out tmp/static-smoke.jsonl --truncate --process-attribution=false --metrics cpu,mem`
    - `./bin/epagent analyze --in tmp/static-smoke.jsonl --format json --window 5 --threshold 10 --static-threshold mem=1 --min-severity low > tmp/static-analyze.json`
    - `./bin/epagent report --in tmp/static-smoke.jsonl --out tmp/static-report.md --window 5 --threshold 10 --static-threshold mem=1 --min-severity low`
    - `./bin/epagent watch --duration 3s --interval 1s --metrics mem --process-attribution=false --sink stdout --min-severity low --threshold 10 --static-threshold mem=1 > tmp/watch-static.ndjson`
- Commit: `76f95ad`
- Confidence: high
- Trust label: trusted (local code/tests)
- Additional market context used for prioritization:
  - Datadog metric/anomaly monitor docs, Telegraf jitter/interval docs, OTel hostmetrics interval docs, Metricbeat period/tag docs.
  - Trust label: untrusted (external docs/web); used for product-priority heuristics only.

### 2026-02-10 - Add selftest, output redaction, and safe overwrite for sample files
- Decision:
  - Add `epagent selftest` to validate host metric availability and estimate sampling overhead (including process attribution overhead probe).
  - Add `--redact omit|hash` to `analyze`/`report` and `watch` alerts to omit/hash `host_id` and labels for sharing.
  - Add `collect --truncate` and `watch --out ... --truncate` to overwrite sample files instead of appending.
- Why: Production readiness needs (1) a quick host readiness check, (2) a safe sharing mode for outputs, and (3) a way to avoid silently appending multiple runs into one sample file.
- Evidence:
  - Code: `cmd/epagent/main.go`, `internal/selftest/selftest.go`, `internal/redact/redact.go`, `internal/storage/storage.go`
  - Tests: `cmd/epagent/main_test.go`, `internal/redact/redact_test.go`, `internal/storage/storage_test.go`
  - Local smoke:
    - `./bin/epagent selftest --format text --runs 1 --timeout 2s`
    - `./bin/epagent collect --once --out tmp/trunc-smoke.jsonl --truncate --process-attribution=false --metrics cpu`
    - `./bin/epagent analyze --in tmp/epagent-smoke.jsonl --format json --window 5 --threshold 3 --redact omit`
    - `./bin/epagent report --in tmp/epagent-smoke.jsonl --out - --window 5 --threshold 3 --redact hash`
    - `./bin/epagent watch --duration 2s --interval 1s --process-attribution=false --metrics cpu,mem --sink stdout --min-severity critical --redact omit > tmp/watch-smoke.ndjson`
- Commit: `6ef0b86`
- Confidence: high
- Trust label: verified-local

### 2026-02-10 - Add labels for ingestion routing and metric-family output filtering
- Decision:
  - Add config `labels` and CLI `--label k=v` (repeatable) for `collect`/`watch`, persisted in JSONL samples and propagated to alerts and analysis/report outputs.
  - Add `analyze`/`report` `--metric cpu|mem|disk|net` (repeatable) to filter output by metric family without re-collecting.
- Why: Real-world usage commonly needs (1) stable dimensions for multi-host/multi-service routing and (2) the ability to focus reports/alerts on a subset of metric families during an incident.
- Evidence:
  - Code: `internal/config/config.go`, `cmd/epagent/main.go`, `internal/collector/collector.go`, `internal/anomaly/anomaly.go`, `internal/alert/alert.go`, `internal/watch/engine.go`, `internal/report/report.go`, `internal/report/metric_family_filter.go`
  - Tests: `internal/config/config_test.go`, `internal/report/report_test.go`, `internal/report/metric_family_filter_test.go`, `internal/watch/engine_test.go`, `cmd/epagent/main_test.go`
  - Local smoke:
    - `./bin/epagent collect --once --out tmp/smoke-labels.jsonl --process-attribution=false --metrics cpu,mem --host-id smoke-host --label env=dev --label service=smoke`
    - `./bin/epagent analyze --in tmp/smoke-labels.jsonl --format json --window 5 --threshold 3`
    - `./bin/epagent analyze --in tmp/smoke-labels.jsonl --format json --window 5 --threshold 3 --metric cpu`
    - `./bin/epagent analyze --in tmp/anom-labels.jsonl --format ndjson --sink stdout --window 5 --threshold 2.5 --metric cpu`
    - `./bin/epagent report --in tmp/anom-labels.jsonl --out - --window 5 --threshold 2.5 --metric cpu`
- Commit: `bdf23c7`
- Confidence: high
- Trust label: verified-local

### 2026-02-09 - Add host ID override and time-window filtering for incident-focused analysis
- Decision:
  - Add `collect`/`watch` `--host-id` to override `host_id` without editing config.
  - Add `analyze`/`report` `--since` and `--until` (RFC3339) to filter samples by an incident time window.
  - Add `analyze`/`report` `--last <duration>` convenience filtering relative to the last sample timestamp in the file.
- Why: Operations workflows often need (1) consistent host identifiers for multi-host ingestion and (2) quick scoping of analysis/reporting to an incident window without pre-slicing JSONL files.
- Evidence:
  - Code: `cmd/epagent/main.go`, `internal/report/sample_filter.go`
  - Tests: `cmd/epagent/main_test.go`, `internal/report/sample_filter_test.go`
  - Local smoke:
    - `./bin/epagent collect --once --out tmp/smoke-time.jsonl --process-attribution=false --metrics cpu,mem --host-id smoke-host`
    - `./bin/epagent analyze --in tmp/smoke-time.jsonl --format json --window 5 --threshold 3 --since 2000-01-01T00:00:00Z --until 2100-01-01T00:00:00Z`
    - `./bin/epagent analyze --in tmp/smoke-last.jsonl --format json --window 5 --threshold 3 --last 2s`
- Commit: `846221d`
- Confidence: high
- Trust label: verified-local

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
- `gh issue list --state open --limit 100 --json number,title,author,createdAt,labels,url` (pass; no open issues from `sarveshkapre` or trusted bots)
- `gh run list --limit 12 --branch main --json databaseId,workflowName,status,conclusion,headSha,createdAt,url` (pass; prior runs healthy before implementation)
- `make check` (pass; static-threshold changes + tests; warnings from `github.com/shoenig/go-m1cpu` on Apple Silicon)
- `./bin/epagent collect --duration 6s --interval 1s --out tmp/static-smoke.jsonl --truncate --process-attribution=false --metrics cpu,mem` (pass)
- `./bin/epagent analyze --in tmp/static-smoke.jsonl --format json --window 5 --threshold 10 --static-threshold mem=1 --min-severity low > tmp/static-analyze.json` (pass)
- `./bin/epagent report --in tmp/static-smoke.jsonl --out tmp/static-report.md --window 5 --threshold 10 --static-threshold mem=1 --min-severity low` (pass)
- `./bin/epagent watch --duration 3s --interval 1s --metrics mem --process-attribution=false --sink stdout --min-severity low --threshold 10 --static-threshold mem=1 > tmp/watch-static.ndjson` (pass)
- `rg -n '"rule_type"|"threshold"' tmp/static-analyze.json | head -n 6` (pass; confirms static rule metadata in JSON output)
- `rg -n 'static threshold' tmp/static-report.md | head -n 3` (pass; confirms markdown static-threshold wording)
- `head -n 1 tmp/watch-static.ndjson` (pass; confirms alert includes `rule_type=static_threshold` and `threshold`)
- `gh run view 21895242856 --json status,conclusion,headSha,url,workflowName` (pass; `ci` completed `success` for `76f95ad`)
- `gh run view 21895242862 --json status,conclusion,headSha,url,workflowName` (pass; `secret-scan` completed `success` for `76f95ad`)
- `gh run view 21895242864 --json status,conclusion,headSha,url,workflowName` (in_progress at last check for `76f95ad`)
- `make check` (pass)
- `make check` (pass; warnings from `github.com/shoenig/go-m1cpu` on Apple Silicon)
- `make check` (pass; warnings from `github.com/shoenig/go-m1cpu` on Apple Silicon)
- `./bin/epagent selftest --format text --runs 1 --timeout 2s` (pass)
- `f=$(mktemp tmp/trunc-smoke.XXXXXX.jsonl) && ./bin/epagent collect --once --out "$f" --truncate --process-attribution=false --metrics cpu && wc -l "$f" && ./bin/epagent collect --once --out "$f" --process-attribution=false --metrics cpu && wc -l "$f" && ./bin/epagent collect --once --out "$f" --truncate --process-attribution=false --metrics cpu && wc -l "$f"` (pass; line counts 1 -> 2 -> 1)
- `./bin/epagent collect --once --out tmp/epagent-smoke.jsonl --truncate --process-attribution=false --metrics cpu,mem --host-id smoke-host --label env=dev --label service=smoke` (pass)
- `./bin/epagent analyze --in tmp/epagent-smoke.jsonl --format json --window 5 --threshold 3 --redact omit` (pass; host_id/labels omitted)
- `./bin/epagent report --in tmp/epagent-smoke.jsonl --out - --window 5 --threshold 3 --redact hash` (pass; host_id/labels hashed)
- `./bin/epagent watch --duration 2s --interval 1s --process-attribution=false --metrics cpu,mem --sink stdout --min-severity critical --redact omit > tmp/watch-smoke.ndjson` (pass)
- `gh run list -L 6 --branch main` (pass; external; shows `ci`, `secret-scan`, `codeql` succeeded for latest `main` push)
- `./bin/epagent collect --once --out tmp/smoke-labels.jsonl --process-attribution=false --metrics cpu,mem --host-id smoke-host --label env=dev --label service=smoke` (pass; JSONL includes labels)
- `head -n 1 tmp/smoke-labels.jsonl` (pass; shows `"labels":{...}`)
- `./bin/epagent analyze --in tmp/smoke-labels.jsonl --format json --window 5 --threshold 3` (pass; includes `labels` in JSON output)
- `./bin/epagent analyze --in tmp/smoke-labels.jsonl --format json --window 5 --threshold 3 --metric cpu` (pass; filtered output excludes mem/disk/net)
- `./bin/epagent analyze --in tmp/anom-labels.jsonl --format ndjson --sink stdout --window 5 --threshold 2.5 --metric cpu | head -n 1` (pass; emitted 1 alert with `labels`)
- `./bin/epagent report --in tmp/anom-labels.jsonl --out - --window 5 --threshold 2.5 --metric cpu | head -n 30` (pass; includes labels + filtered baselines)
- `gh run view 21845904514 --json status,conclusion,headSha` (pass; `ci` completed `success` for `bdf23c7`)
- `gh run view 21845904526 --json status,conclusion,headSha` (pass; `codeql` completed `success` for `bdf23c7`)
- `gh run view 21845912802 --json status,conclusion,headSha` (pass; `ci` completed `success` for `cdebc80`)
- `gh run view 21845912789 --json status,conclusion,headSha` (pass; `codeql` completed `success` for `cdebc80`)
- `gh run view 21845956694 --json status,conclusion,headSha` (pass; `ci` completed `success` for `f550a60`)
- `gh run view 21845960484 --json status,conclusion,headSha` (pass; `codeql` completed `success` for `f550a60`)
- `./bin/epagent collect --once --out tmp/smoke-time.jsonl --process-attribution=false --metrics cpu,mem --host-id smoke-host` (pass; JSONL includes host_id=smoke-host)
- `./bin/epagent analyze --in tmp/smoke-time.jsonl --format json --window 5 --threshold 3 --since 2000-01-01T00:00:00Z --until 2100-01-01T00:00:00Z` (pass)
- `./bin/epagent report --in tmp/smoke-time.jsonl --out - --window 5 --threshold 3 --since 2000-01-01T00:00:00Z --until 2100-01-01T00:00:00Z` (pass)
- `./bin/epagent analyze --in tmp/smoke-last.jsonl --format json --window 5 --threshold 3 --last 2s` (pass; samples=3)
- `./bin/epagent analyze --in tmp/smoke-time.jsonl --format text --since 2026-02-09T19:58:24.322408Z` (pass; accepts fractional seconds RFC3339)
- `gh run list -L 3 --branch main` (pass; `ci`, `secret-scan`, `codeql` succeeded for latest push)
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
