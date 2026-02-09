# Endpoint Perf Agent

Cross-platform endpoint performance agent that samples CPU/memory/disk/network metrics, detects anomalies from a rolling baseline, and generates human-readable explanations.

## Why
- Lightweight local-first agent for catching regressions or unexpected spikes.
- Runs without accounts or external services.
- Generates explainers to speed up triage.

## Features
- CPU, memory, disk, and network sampling (cross-platform via gopsutil).
- Metric family allow-listing (cpu/mem/disk/net) to tune overhead and reduce noise.
- Per-sample top CPU and top memory process attribution for triage context.
- Rolling z-score anomaly detection with severity levels.
- JSONL storage for easy ingestion.
- Markdown/JSON analysis output with anomaly timestamps, process context, and baseline summaries.

## Quickstart
```bash
make setup
make build

# collect samples for 60 seconds
./bin/epagent collect --duration 60s

# stream alerts for 60 seconds (NDJSON to stdout)
./bin/epagent watch --duration 60s --min-severity medium --sink stdout

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
  "enabled_metrics": ["cpu", "mem", "disk", "net"],
  "window_size": 30,
  "zscore_threshold": 3.0,
  "output_path": "data/metrics.jsonl",
  "host_id": "laptop-01",
  "process_attribution": true
}
```

## Commands
```bash
epagent collect --once
epagent collect --duration 60s --metrics cpu,mem
epagent watch --min-severity high --sink stdout
epagent watch --duration 60s --metrics cpu,mem --sink syslog
epagent analyze --in data/metrics.jsonl --window 30 --threshold 3
epagent analyze --in data/metrics.jsonl --format json  # includes baselines
epagent analyze --in data/metrics.jsonl --min-severity high --top 10
epagent report --out endpoint-perf-report.md
epagent report --min-severity medium --top 20 --out -
epagent report --out -
```

## Docker
Not applicable. This agent relies on host-level metrics that are not reliably available inside containers without elevated privileges.

## Security
- Local-only by default; no network calls.
- JSONL output can contain host identifiers; handle as sensitive.

## License
MIT.
