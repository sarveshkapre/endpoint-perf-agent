# INCIDENTS

## 2026-02-09 - Smoke flow race in automation run
- Summary: A smoke verification attempt ran `collect`, `analyze`, and `report` in parallel; `analyze` and `report` failed because the sample file had not been created yet.
- Impact: False-negative verification signal during maintenance session (no product regression shipped).
- Root cause: Dependent commands were executed concurrently.
- Detection: CLI error `open tmp/smoke-metrics.jsonl: no such file or directory`.
- Resolution: Re-ran smoke commands sequentially with `collect` first.
- Prevention rules:
  - Execute dependent smoke checks in strict sequence.
  - Reserve parallel execution for independent reads/checks only.
- Status: resolved
