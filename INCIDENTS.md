# INCIDENTS

## 2026-02-09 - `watch` panic when `--out` is unset
- Summary: `epagent watch` could panic when run without `--out`, even though sample writing is meant to be optional.
- Impact: High. Default-ish invocation (`--out` omitted) could crash the process instead of emitting alerts.
- Root cause: A typed `nil` (`*storage.Writer`) was assigned to an interface field (`Runner.Writer`), so the `!= nil` guard passed and the subsequent method call panicked.
- Detection: Local smoke run: `./bin/epagent watch --duration 2s --interval 1s --sink stdout --min-severity critical` panicked with `invalid memory address or nil pointer dereference`.
- Resolution:
  - Ensure `cmd/epagent` passes a `nil` interface when `--out` is unset.
  - Harden `internal/watch.Runner` to treat typed-nil interface values as nil.
  - Added regression test for typed-nil writer handling.
- Prevention rules:
  - Never assign a typed-nil pointer to an interface unless you intentionally want it to be non-nil.
  - When storing optional interfaces, guard with an `isNilInterface` check (or avoid interfaces for optional pointers).
- Status: resolved

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

### 2026-02-12T20:01:34Z | Codex execution failure
- Date: 2026-02-12T20:01:34Z
- Trigger: Codex execution failure
- Impact: Repo session did not complete cleanly
- Root Cause: codex exec returned a non-zero status
- Fix: Captured failure logs and kept repository in a recoverable state
- Prevention Rule: Re-run with same pass context and inspect pass log before retrying
- Evidence: pass_log=logs/20260212-101456-endpoint-perf-agent-cycle-2.log
- Commit: pending
- Confidence: medium

### 2026-02-12T20:05:01Z | Codex execution failure
- Date: 2026-02-12T20:05:01Z
- Trigger: Codex execution failure
- Impact: Repo session did not complete cleanly
- Root Cause: codex exec returned a non-zero status
- Fix: Captured failure logs and kept repository in a recoverable state
- Prevention Rule: Re-run with same pass context and inspect pass log before retrying
- Evidence: pass_log=logs/20260212-101456-endpoint-perf-agent-cycle-3.log
- Commit: pending
- Confidence: medium

### 2026-02-12T20:08:31Z | Codex execution failure
- Date: 2026-02-12T20:08:31Z
- Trigger: Codex execution failure
- Impact: Repo session did not complete cleanly
- Root Cause: codex exec returned a non-zero status
- Fix: Captured failure logs and kept repository in a recoverable state
- Prevention Rule: Re-run with same pass context and inspect pass log before retrying
- Evidence: pass_log=logs/20260212-101456-endpoint-perf-agent-cycle-4.log
- Commit: pending
- Confidence: medium

### 2026-02-12T20:11:58Z | Codex execution failure
- Date: 2026-02-12T20:11:58Z
- Trigger: Codex execution failure
- Impact: Repo session did not complete cleanly
- Root Cause: codex exec returned a non-zero status
- Fix: Captured failure logs and kept repository in a recoverable state
- Prevention Rule: Re-run with same pass context and inspect pass log before retrying
- Evidence: pass_log=logs/20260212-101456-endpoint-perf-agent-cycle-5.log
- Commit: pending
- Confidence: medium

### 2026-02-12T20:15:27Z | Codex execution failure
- Date: 2026-02-12T20:15:27Z
- Trigger: Codex execution failure
- Impact: Repo session did not complete cleanly
- Root Cause: codex exec returned a non-zero status
- Fix: Captured failure logs and kept repository in a recoverable state
- Prevention Rule: Re-run with same pass context and inspect pass log before retrying
- Evidence: pass_log=logs/20260212-101456-endpoint-perf-agent-cycle-6.log
- Commit: pending
- Confidence: medium

### 2026-02-12T20:18:58Z | Codex execution failure
- Date: 2026-02-12T20:18:58Z
- Trigger: Codex execution failure
- Impact: Repo session did not complete cleanly
- Root Cause: codex exec returned a non-zero status
- Fix: Captured failure logs and kept repository in a recoverable state
- Prevention Rule: Re-run with same pass context and inspect pass log before retrying
- Evidence: pass_log=logs/20260212-101456-endpoint-perf-agent-cycle-7.log
- Commit: pending
- Confidence: medium

### 2026-02-12T20:22:24Z | Codex execution failure
- Date: 2026-02-12T20:22:24Z
- Trigger: Codex execution failure
- Impact: Repo session did not complete cleanly
- Root Cause: codex exec returned a non-zero status
- Fix: Captured failure logs and kept repository in a recoverable state
- Prevention Rule: Re-run with same pass context and inspect pass log before retrying
- Evidence: pass_log=logs/20260212-101456-endpoint-perf-agent-cycle-8.log
- Commit: pending
- Confidence: medium

### 2026-02-12T20:25:57Z | Codex execution failure
- Date: 2026-02-12T20:25:57Z
- Trigger: Codex execution failure
- Impact: Repo session did not complete cleanly
- Root Cause: codex exec returned a non-zero status
- Fix: Captured failure logs and kept repository in a recoverable state
- Prevention Rule: Re-run with same pass context and inspect pass log before retrying
- Evidence: pass_log=logs/20260212-101456-endpoint-perf-agent-cycle-9.log
- Commit: pending
- Confidence: medium

### 2026-02-12T20:29:36Z | Codex execution failure
- Date: 2026-02-12T20:29:36Z
- Trigger: Codex execution failure
- Impact: Repo session did not complete cleanly
- Root Cause: codex exec returned a non-zero status
- Fix: Captured failure logs and kept repository in a recoverable state
- Prevention Rule: Re-run with same pass context and inspect pass log before retrying
- Evidence: pass_log=logs/20260212-101456-endpoint-perf-agent-cycle-10.log
- Commit: pending
- Confidence: medium

### 2026-02-12T20:33:01Z | Codex execution failure
- Date: 2026-02-12T20:33:01Z
- Trigger: Codex execution failure
- Impact: Repo session did not complete cleanly
- Root Cause: codex exec returned a non-zero status
- Fix: Captured failure logs and kept repository in a recoverable state
- Prevention Rule: Re-run with same pass context and inspect pass log before retrying
- Evidence: pass_log=logs/20260212-101456-endpoint-perf-agent-cycle-11.log
- Commit: pending
- Confidence: medium

### 2026-02-12T20:36:31Z | Codex execution failure
- Date: 2026-02-12T20:36:31Z
- Trigger: Codex execution failure
- Impact: Repo session did not complete cleanly
- Root Cause: codex exec returned a non-zero status
- Fix: Captured failure logs and kept repository in a recoverable state
- Prevention Rule: Re-run with same pass context and inspect pass log before retrying
- Evidence: pass_log=logs/20260212-101456-endpoint-perf-agent-cycle-12.log
- Commit: pending
- Confidence: medium

### 2026-02-12T20:39:58Z | Codex execution failure
- Date: 2026-02-12T20:39:58Z
- Trigger: Codex execution failure
- Impact: Repo session did not complete cleanly
- Root Cause: codex exec returned a non-zero status
- Fix: Captured failure logs and kept repository in a recoverable state
- Prevention Rule: Re-run with same pass context and inspect pass log before retrying
- Evidence: pass_log=logs/20260212-101456-endpoint-perf-agent-cycle-13.log
- Commit: pending
- Confidence: medium

### 2026-02-12T20:43:26Z | Codex execution failure
- Date: 2026-02-12T20:43:26Z
- Trigger: Codex execution failure
- Impact: Repo session did not complete cleanly
- Root Cause: codex exec returned a non-zero status
- Fix: Captured failure logs and kept repository in a recoverable state
- Prevention Rule: Re-run with same pass context and inspect pass log before retrying
- Evidence: pass_log=logs/20260212-101456-endpoint-perf-agent-cycle-14.log
- Commit: pending
- Confidence: medium

### 2026-02-12T20:47:00Z | Codex execution failure
- Date: 2026-02-12T20:47:00Z
- Trigger: Codex execution failure
- Impact: Repo session did not complete cleanly
- Root Cause: codex exec returned a non-zero status
- Fix: Captured failure logs and kept repository in a recoverable state
- Prevention Rule: Re-run with same pass context and inspect pass log before retrying
- Evidence: pass_log=logs/20260212-101456-endpoint-perf-agent-cycle-15.log
- Commit: pending
- Confidence: medium

### 2026-02-12T20:50:31Z | Codex execution failure
- Date: 2026-02-12T20:50:31Z
- Trigger: Codex execution failure
- Impact: Repo session did not complete cleanly
- Root Cause: codex exec returned a non-zero status
- Fix: Captured failure logs and kept repository in a recoverable state
- Prevention Rule: Re-run with same pass context and inspect pass log before retrying
- Evidence: pass_log=logs/20260212-101456-endpoint-perf-agent-cycle-16.log
- Commit: pending
- Confidence: medium

### 2026-02-12T20:54:03Z | Codex execution failure
- Date: 2026-02-12T20:54:03Z
- Trigger: Codex execution failure
- Impact: Repo session did not complete cleanly
- Root Cause: codex exec returned a non-zero status
- Fix: Captured failure logs and kept repository in a recoverable state
- Prevention Rule: Re-run with same pass context and inspect pass log before retrying
- Evidence: pass_log=logs/20260212-101456-endpoint-perf-agent-cycle-17.log
- Commit: pending
- Confidence: medium

### 2026-02-12T20:57:35Z | Codex execution failure
- Date: 2026-02-12T20:57:35Z
- Trigger: Codex execution failure
- Impact: Repo session did not complete cleanly
- Root Cause: codex exec returned a non-zero status
- Fix: Captured failure logs and kept repository in a recoverable state
- Prevention Rule: Re-run with same pass context and inspect pass log before retrying
- Evidence: pass_log=logs/20260212-101456-endpoint-perf-agent-cycle-18.log
- Commit: pending
- Confidence: medium

### 2026-02-12T21:01:01Z | Codex execution failure
- Date: 2026-02-12T21:01:01Z
- Trigger: Codex execution failure
- Impact: Repo session did not complete cleanly
- Root Cause: codex exec returned a non-zero status
- Fix: Captured failure logs and kept repository in a recoverable state
- Prevention Rule: Re-run with same pass context and inspect pass log before retrying
- Evidence: pass_log=logs/20260212-101456-endpoint-perf-agent-cycle-19.log
- Commit: pending
- Confidence: medium

### 2026-02-12T21:04:29Z | Codex execution failure
- Date: 2026-02-12T21:04:29Z
- Trigger: Codex execution failure
- Impact: Repo session did not complete cleanly
- Root Cause: codex exec returned a non-zero status
- Fix: Captured failure logs and kept repository in a recoverable state
- Prevention Rule: Re-run with same pass context and inspect pass log before retrying
- Evidence: pass_log=logs/20260212-101456-endpoint-perf-agent-cycle-20.log
- Commit: pending
- Confidence: medium

### 2026-02-12T21:08:02Z | Codex execution failure
- Date: 2026-02-12T21:08:02Z
- Trigger: Codex execution failure
- Impact: Repo session did not complete cleanly
- Root Cause: codex exec returned a non-zero status
- Fix: Captured failure logs and kept repository in a recoverable state
- Prevention Rule: Re-run with same pass context and inspect pass log before retrying
- Evidence: pass_log=logs/20260212-101456-endpoint-perf-agent-cycle-21.log
- Commit: pending
- Confidence: medium

### 2026-02-12T21:11:34Z | Codex execution failure
- Date: 2026-02-12T21:11:34Z
- Trigger: Codex execution failure
- Impact: Repo session did not complete cleanly
- Root Cause: codex exec returned a non-zero status
- Fix: Captured failure logs and kept repository in a recoverable state
- Prevention Rule: Re-run with same pass context and inspect pass log before retrying
- Evidence: pass_log=logs/20260212-101456-endpoint-perf-agent-cycle-22.log
- Commit: pending
- Confidence: medium

### 2026-02-12T21:15:05Z | Codex execution failure
- Date: 2026-02-12T21:15:05Z
- Trigger: Codex execution failure
- Impact: Repo session did not complete cleanly
- Root Cause: codex exec returned a non-zero status
- Fix: Captured failure logs and kept repository in a recoverable state
- Prevention Rule: Re-run with same pass context and inspect pass log before retrying
- Evidence: pass_log=logs/20260212-101456-endpoint-perf-agent-cycle-23.log
- Commit: pending
- Confidence: medium

### 2026-02-12T21:18:34Z | Codex execution failure
- Date: 2026-02-12T21:18:34Z
- Trigger: Codex execution failure
- Impact: Repo session did not complete cleanly
- Root Cause: codex exec returned a non-zero status
- Fix: Captured failure logs and kept repository in a recoverable state
- Prevention Rule: Re-run with same pass context and inspect pass log before retrying
- Evidence: pass_log=logs/20260212-101456-endpoint-perf-agent-cycle-24.log
- Commit: pending
- Confidence: medium

### 2026-02-12T21:21:51Z | Codex execution failure
- Date: 2026-02-12T21:21:51Z
- Trigger: Codex execution failure
- Impact: Repo session did not complete cleanly
- Root Cause: codex exec returned a non-zero status
- Fix: Captured failure logs and kept repository in a recoverable state
- Prevention Rule: Re-run with same pass context and inspect pass log before retrying
- Evidence: pass_log=logs/20260212-101456-endpoint-perf-agent-cycle-25.log
- Commit: pending
- Confidence: medium

### 2026-02-12T21:25:05Z | Codex execution failure
- Date: 2026-02-12T21:25:05Z
- Trigger: Codex execution failure
- Impact: Repo session did not complete cleanly
- Root Cause: codex exec returned a non-zero status
- Fix: Captured failure logs and kept repository in a recoverable state
- Prevention Rule: Re-run with same pass context and inspect pass log before retrying
- Evidence: pass_log=logs/20260212-101456-endpoint-perf-agent-cycle-26.log
- Commit: pending
- Confidence: medium

### 2026-02-12T21:28:25Z | Codex execution failure
- Date: 2026-02-12T21:28:25Z
- Trigger: Codex execution failure
- Impact: Repo session did not complete cleanly
- Root Cause: codex exec returned a non-zero status
- Fix: Captured failure logs and kept repository in a recoverable state
- Prevention Rule: Re-run with same pass context and inspect pass log before retrying
- Evidence: pass_log=logs/20260212-101456-endpoint-perf-agent-cycle-27.log
- Commit: pending
- Confidence: medium

### 2026-02-12T21:31:47Z | Codex execution failure
- Date: 2026-02-12T21:31:47Z
- Trigger: Codex execution failure
- Impact: Repo session did not complete cleanly
- Root Cause: codex exec returned a non-zero status
- Fix: Captured failure logs and kept repository in a recoverable state
- Prevention Rule: Re-run with same pass context and inspect pass log before retrying
- Evidence: pass_log=logs/20260212-101456-endpoint-perf-agent-cycle-28.log
- Commit: pending
- Confidence: medium

### 2026-02-12T21:35:15Z | Codex execution failure
- Date: 2026-02-12T21:35:15Z
- Trigger: Codex execution failure
- Impact: Repo session did not complete cleanly
- Root Cause: codex exec returned a non-zero status
- Fix: Captured failure logs and kept repository in a recoverable state
- Prevention Rule: Re-run with same pass context and inspect pass log before retrying
- Evidence: pass_log=logs/20260212-101456-endpoint-perf-agent-cycle-29.log
- Commit: pending
- Confidence: medium
