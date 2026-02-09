# CHANGELOG

## Unreleased
- Implemented sampling, anomaly detection, and report generation.
- Added CLI commands and JSONL storage.
- `analyze --format json` for machine-readable output.
- `report --out -` to write Markdown to stdout.
- Baseline summaries (mean/stddev/min/max) in Markdown report and JSON output.
- Robust sample handling (timestamp sorting, tolerate blank lines) and clearer explanations.
- `--min-severity` and `--top` filtering for `analyze` and `report`.
- Added per-sample top-process attribution (`top_cpu_process`, `top_mem_process`) in collected JSONL.
- Added anomaly context in analysis outputs (timestamp + top CPU/memory process details).
- Normalized analysis window/threshold values to detector-safe minimums in output (`window>=5`, `threshold>0`).
- Added stricter CLI validation for negative `--window`, `--threshold`, and `--top` values.
- Improved JSONL parse errors with line numbers for faster corruption diagnosis.
- Added `collect --process-attribution=false` and config `process_attribution` to disable per-sample process scans when overhead is a concern.
