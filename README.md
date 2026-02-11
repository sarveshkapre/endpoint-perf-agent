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
- Optional static-threshold alert rules for absolute ceilings (e.g., CPU > 85%).
- JSONL storage for easy ingestion.
- Markdown/JSON analysis output with anomaly timestamps, process context, and baseline summaries.

## Quickstart
```bash
make setup
make build

# collect samples for 60 seconds
./bin/epagent collect --duration 60s

# validate host permissions / overhead
./bin/epagent selftest --format text

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
  "static_thresholds": {
    "cpu_percent": 85,
    "mem_used_percent": 90
  },
  "output_path": "data/metrics.jsonl",
  "host_id": "laptop-01",
  "labels": { "env": "dev", "service": "api" },
  "process_attribution": true
}
```
You can override `host_id` at runtime with `--host-id` on `collect` and `watch`.
You can add/override labels at runtime with `--label k=v` (repeatable) on `collect` and `watch`.

## Commands
```bash
epagent collect --once
epagent collect --once --out -
epagent collect --duration 60s --out data/metrics.jsonl --truncate
epagent collect --once --host-id laptop-01
epagent collect --duration 60s --label env=prod --label service=api
epagent collect --duration 60s --metrics cpu,mem
epagent watch --min-severity high --sink stdout
epagent watch --duration 60s --host-id laptop-01 --metrics cpu,mem --sink stdout
epagent watch --duration 60s --label env=prod --label service=api --sink stdout
epagent watch --duration 60s --metrics cpu,mem --sink syslog
epagent watch --duration 60s --metrics cpu,mem --static-threshold cpu=85 --sink stdout
epagent analyze --in data/metrics.jsonl --window 30 --threshold 3
epagent analyze --in data/metrics.jsonl --window 30 --threshold 10 --static-threshold mem=90
epagent analyze --in data/metrics.jsonl --format json  # includes baselines
epagent analyze --in data/metrics.jsonl --format ndjson --sink stdout  # one alert per line
epagent analyze --in data/metrics.jsonl --metric cpu --metric net  # filter output by metric family
epagent analyze --in data/metrics.jsonl --format json --redact omit  # omit host_id/labels for sharing
epagent analyze --in data/metrics.jsonl --since 2026-02-09T00:00:00Z --until 2026-02-09T00:10:00Z
epagent analyze --in data/metrics.jsonl --last 10m
epagent analyze --in data/metrics.jsonl --min-severity high --top 10
epagent report --out endpoint-perf-report.md
epagent report --min-severity medium --top 20 --out -
epagent report --in data/metrics.jsonl --since 2026-02-09T00:00:00Z --until 2026-02-09T00:10:00Z --out -
epagent report --in data/metrics.jsonl --metric cpu --out -  # filter output by metric family
epagent report --in data/metrics.jsonl --last 10m --out -
epagent report --in data/metrics.jsonl --window 30 --threshold 10 --static-threshold disk_used_percent=80 --out -
epagent report --out -
epagent report --in data/metrics.jsonl --out - --redact hash  # hash host_id/labels for sharing
epagent selftest --format json --runs 3 --timeout 2s
```

## Docker
Not applicable. This agent relies on host-level metrics that are not reliably available inside containers without elevated privileges.

## Security
- Local-only by default; no network calls.
- JSONL output can contain host identifiers; handle as sensitive.

## License
MIT.
