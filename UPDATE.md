# Update (2026-02-10)

## Shipped
- Added optional labels for multi-host/multi-service ingestion:
  - config `labels` and `collect`/`watch` `--label k=v` (repeatable)
  - labels propagate into JSONL samples, alerts, and analysis/report outputs
- Added `analyze`/`report` `--metric cpu|mem|disk|net` (repeatable) to filter output by metric family.

## Verify
- `make check`
- `./bin/epagent collect --once --out tmp/smoke-labels.jsonl --process-attribution=false --metrics cpu,mem --host-id smoke-host --label env=dev --label service=smoke`
- `./bin/epagent analyze --in tmp/smoke-labels.jsonl --format json --window 5 --threshold 3`
- `./bin/epagent analyze --in tmp/smoke-labels.jsonl --format json --window 5 --threshold 3 --metric cpu`
- `./bin/epagent analyze --in tmp/anom-labels.jsonl --format ndjson --sink stdout --window 5 --threshold 2.5 --metric cpu`
- `./bin/epagent report --in tmp/anom-labels.jsonl --out - --window 5 --threshold 2.5 --metric cpu`

## Notes
- No external API integration was changed; external integration smoke checks are not applicable for this update.

# Update (2026-02-09)

## Shipped
- Added per-sample process attribution during collection:
  - `top_cpu_process` (pid/name/cpu/rss)
  - `top_mem_process` (pid/name/cpu/rss)
- Added richer anomaly context in `analyze` and `report` outputs:
  - anomaly timestamp
  - top CPU and top memory process details at detection time
- Hardened runtime parameter handling:
  - normalize analysis params to detector-safe values (`window>=5`, `threshold>0`)
  - reject negative `--window`, `--threshold`, and `--top`
  - accept case-insensitive severity filters for `--min-severity`
- Improved JSONL corruption diagnostics with line-numbered parse errors.

## Verify
- `make check`
- `./bin/epagent collect --once --out tmp/smoke-metrics.jsonl`
- `./bin/epagent analyze --in tmp/smoke-metrics.jsonl --format json`
- `./bin/epagent report --in tmp/smoke-metrics.jsonl --out -`
- `./bin/epagent analyze --in tmp/smoke-metrics.jsonl --top -1` (expects non-zero exit + validation error)

## Notes
- No external API integration was changed; external integration smoke checks are not applicable for this update.

# Update (2026-02-01)

## Shipped
- Added machine-readable analysis output: `epagent analyze --format json`.
- Added report-to-stdout: `epagent report --out -`.
- Added per-metric baseline stats (mean/stddev/min/max) to analysis outputs (Markdown + JSON).
- Made analysis/reporting more robust:
  - Sorts samples by timestamp before analyzing.
  - Skips blank lines and supports larger JSONL lines.
  - Direction-aware anomaly explanations (spike vs drop).
- Added output controls:
  - `--min-severity low|medium|high|critical`
  - `--top N` (limit by absolute z-score)

## Verify
- `make check`
- `make build`
- `./bin/epagent collect --duration 30s`
- `./bin/epagent analyze --format json --min-severity medium --top 10`
- `./bin/epagent report --min-severity medium --top 10 --out -`

## Notes
- No PRs were created/updated; changes are committed directly to `main`.
