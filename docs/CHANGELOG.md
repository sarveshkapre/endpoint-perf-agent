# CHANGELOG

## Unreleased
- Implemented sampling, anomaly detection, and report generation.
- Added CLI commands and JSONL storage.
- `analyze --format json` for machine-readable output.
- `report --out -` to write Markdown to stdout.
- `collect`/`watch` `--host-id` to override `host_id` without editing config.
- Added config `labels` and `collect`/`watch` `--label k=v` (repeatable) to tag samples for multi-host and multi-service ingestion; labels propagate into alerts and reports.
- `analyze`/`report` `--since` and `--until` to restrict analysis to a time window.
- `analyze`/`report` `--last <duration>` to quickly focus on the tail end of a sample file.
- Added `analyze`/`report` `--metric cpu|mem|disk|net` (repeatable) to filter output by metric family.
- Baseline summaries (mean/stddev/min/max) in Markdown report and JSON output.
- Robust sample handling (timestamp sorting, tolerate blank lines) and clearer explanations.
- `--min-severity` and `--top` filtering for `analyze` and `report`.
- Added metric family allow-listing (`enabled_metrics` / `--metrics`) to tune overhead by enabling only cpu/mem/disk/net as needed.
- Added `collect --out -` to stream JSONL samples to stdout.
- Added `analyze --format ndjson` and `--sink stdout|syslog` to emit one JSON alert per anomaly for integration into log pipelines.
- Added per-sample top-process attribution (`top_cpu_process`, `top_mem_process`) in collected JSONL.
- Added anomaly context in analysis outputs (timestamp + top CPU/memory process details).
- Normalized analysis window/threshold values to detector-safe minimums in output (`window>=5`, `threshold>0`).
- Added stricter CLI validation for negative `--window`, `--threshold`, and `--top` values.
- Improved JSONL parse errors with line numbers for faster corruption diagnosis.
- Added `collect --process-attribution=false` and config `process_attribution` to disable per-sample process scans when overhead is a concern.
- Added `watch` command to stream anomaly alerts to stdout (NDJSON) or syslog, with per-metric cooldown/dedupe.
- Added `selftest` command to validate host metric availability and estimate collection overhead.
- Added `--redact omit|hash` for `analyze`/`report` outputs and `watch` alerts to omit/hash `host_id` and labels for sharing.
- Added `collect --truncate` and `watch --out ... --truncate` to overwrite sample files instead of appending.
- Added configurable static-threshold rules via config `static_thresholds` and CLI `--static-threshold metric=value` for `watch`/`analyze`/`report`.
- Added anomaly/alert rule metadata (`rule_type`, `threshold`) so NDJSON/JSON outputs can distinguish z-score and static-threshold triggers.
