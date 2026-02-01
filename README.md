# Endpoint Perf Agent

Cross-platform endpoint performance agent that samples CPU/memory/disk/network metrics, detects anomalies from a rolling baseline, and generates human-readable explanations.

## Why
- Lightweight local-first agent for catching regressions or unexpected spikes.
- Runs without accounts or external services.
- Generates explainers to speed up triage.

## Features
- CPU, memory, disk, and network sampling (cross-platform via gopsutil).
- Rolling z-score anomaly detection with severity levels.
- JSONL storage for easy ingestion.
- Markdown report generation with explanations.

## Quickstart
```bash
make setup
make build

# collect samples for 60 seconds
./bin/epagent collect --duration 60s

# analyze samples
./bin/epagent analyze

# generate a report
./bin/epagent report --out report.md
```

## Config
Create a JSON config and pass it to `collect` with `--config`.
```json
{
  "interval": "5s",
  "duration": "1m",
  "window_size": 30,
  "zscore_threshold": 3.0,
  "output_path": "data/metrics.jsonl",
  "host_id": "laptop-01"
}
```

## Commands
```bash
epagent collect --once
epagent analyze --in data/metrics.jsonl --window 30 --threshold 3
epagent report --out endpoint-perf-report.md
```

## Docker
Not applicable. This agent relies on host-level metrics that are not reliably available inside containers without elevated privileges.

## Security
- Local-only by default; no network calls.
- JSONL output can contain host identifiers; handle as sensitive.

## License
MIT.
