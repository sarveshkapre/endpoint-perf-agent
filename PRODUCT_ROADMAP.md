# Product Roadmap

## Product Goal
- Keep endpoint-perf-agent production-ready. Current focus: Endpoint Perf Agent. Find the highest-impact pending work, implement it, test it, and push to main.

## Definition Of Done
- Core feature set delivered for primary workflows.
- UI/UX polished for repeated real usage.
- No open critical reliability issues.
- Verification commands pass and are documented.
- Documentation is current and complete.

## Milestones
- M1 Foundation
- M2 Core Features
- M3 Bug Fixing And Refactor
- M4 UI/UX Improvement
- M5 Stabilization And Release Readiness

## Current Milestone
- M2 Core Features

## Brainstorming Queue
- Keep a broad queue of aligned candidates across features, bugs, refactor, UI/UX, docs, and test hardening.

## Pending Features
- Percentile lower-bound rules (drop detection).
- Time-aware baselines (hour/day seasonality).
- SQLite TTL-based retention and compaction.
- Optional OpenTelemetry anomaly export.

## Delivered Features
- 2026-02-17: Sampling jitter (`sampling_jitter`, `collect/watch --jitter`).
- 2026-02-17: Per-metric watch cooldown overrides (`cooldowns`, `--metric-cooldown`).
- 2026-02-17: Percentile-threshold alert rules (`percentile_thresholds`, `--percentile-threshold`).
- 2026-02-17: Optional SQLite sample storage and auto-read support for analyze/report.

## Risks And Blockers
- Track blockers and mitigation plans.
