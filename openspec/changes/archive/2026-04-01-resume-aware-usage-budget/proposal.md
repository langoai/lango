## Why

When a session is resumed (e.g., user reconnects or process restarts), the BudgetPolicy counters reset to zero because they are in-memory only. This means a resumed session effectively gets a fresh budget — turns and delegations consumed before the resume are lost, allowing sessions to exceed their intended limits. Usage continuity across session resumes is essential for reliable budget enforcement.

## What Changes

- Add `Serialize()` and `Restore()` methods to `BudgetPolicy` for persisting/recovering turn and delegation counters
- Expose `*BudgetPolicy` from `initAgentRuntime` so the wiring layer can access it
- Introduce a `budgetRestoringExecutor` wrapper that lazily restores budget state on first use per session
- Register an `OnTurnComplete` callback that persists budget + token usage into `Session.Metadata` after each turn
- Token usage (input/output) from `MetricsCollector.Snapshot().SessionBreakdown` is also persisted for cumulative tracking

## Capabilities

### New Capabilities
- `resume-aware-usage-budget`: Persist and restore budget state (turns, delegations, token usage) across session resumes using existing Session.Metadata storage

### Modified Capabilities

## Impact

- `internal/agentrt/budget.go` — new `Serialize()` / `Restore()` methods
- `internal/app/wiring_agentrt.go` — return type changes to `(turnrunner.Executor, *agentrt.BudgetPolicy)`
- `internal/app/wiring_session_usage.go` — new file: executor wrapper + turn callback wiring
- `internal/app/app.go` — ~3 lines to handle new return, wrap executor, wire callback
- Storage: uses existing `Session.Metadata` (map[string]string) — no schema changes
