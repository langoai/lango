## Why

Post-implementation code review of the exec-safety-followup batch (6 changes) revealed 3 correctness/security bugs: (1) budget serialization used a process-shared instance instead of session-local state, causing stale writes and cross-session contamination; (2) env wrapper parsing didn't skip flags and variable assignments, leaving shell-wrapper bypass paths open; (3) policy event publishing was gated behind hook configuration, disabling observability by default.

## What Changes

- Fix budget persistence to use session-local cumulative state with per-session stats retrieval from `CoordinatingExecutor`, replacing the stale shared-budget serialization
- Fix env wrapper unwrap to properly skip `-i`, `-u NAME`, `-C DIR`, `-S STRING`, `--`, and `NAME=value` assignments before reaching the shell verb
- Decouple policy event publishing from hook event publishing gate so observability works in default single-agent configuration
- Fix `LastRunStats` concurrency: replace single-slot storage with `sync.Map[sessionID → RunStats]` and consume-once semantics

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `exec-policy-evaluator`: env wrapper unwrap now handles flags and variable assignments per POSIX/GNU env syntax
- `resume-aware-usage-budget`: budget persistence uses session-local state instead of process-shared instance

## Impact

- `internal/agentrt/coordinating_executor.go` — `RunStats` stored per-session, `LastRunStatsForSession` replaces `LastRunStats`
- `internal/app/wiring_session_usage.go` — session-local `sessionBudgetState`, no shared budget dependency
- `internal/app/wiring_agentrt.go` — `initAgentRuntime` return reverted to single `Executor`
- `internal/app/app.go` — policyBus gate simplified, wiring adjusted
- `internal/tools/exec/unwrap.go` — `skipEnvArgs` + `looksLikeEnvAssignment` added
