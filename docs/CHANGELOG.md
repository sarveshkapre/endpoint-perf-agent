# CHANGELOG

## Unreleased
- Implemented sampling, anomaly detection, and report generation.
- Added CLI commands and JSONL storage.
- `analyze --format json` for machine-readable output.
- `report --out -` to write Markdown to stdout.
- Robust sample handling (timestamp sorting, tolerate blank lines) and clearer explanations.
- `--min-severity` and `--top` filtering for `analyze` and `report`.
