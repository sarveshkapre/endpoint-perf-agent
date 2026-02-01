# Update (2026-02-01)

## Shipped
- Added machine-readable analysis output: `epagent analyze --format json`.
- Added report-to-stdout: `epagent report --out -`.
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
