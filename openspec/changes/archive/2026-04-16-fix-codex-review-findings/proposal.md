## Why

Codex review (base: dev) found P1×1 + P2×4 issues in the agent orchestration roadmap branch. All five are valid bugs/design flaws. Additionally, the `cli-agent-tools-hooks` spec has a stale 3-arg signature that doesn't match the 4-arg runtime code.

## What Changes

- **Fix 1 (P1)**: Pass real `KnowledgeSaver` from `intelligenceValues.KC.store` into `buildHookRegistry` so runtime knowledge save actually persists
- **Fix 2 (P2)**: Exclude the current bootstrap run from the timing baseline in `BootstrapTimingCheck`
- **Fix 3 (P2)**: Return error from `listRunLedgerWorktrees` on `git worktree list` failure and `scanner.Err()`, report as Warn instead of false Pass
- **Fix 4 (P2)**: Add `firstSeen` timestamp to drift counters so stale cross-window accumulation is rejected, fix `Prune()` key mismatch
- **Fix 5 (P2)**: Add `-u` flag to `git stash push` suggestion so untracked files are included
- **Fix 6 (spec)**: Update `BuildHookRegistry` spec to 4-arg signature matching code

## Capabilities

### New Capabilities
_(none)_

### Modified Capabilities
- `cli-agent-tools-hooks`: Fix `BuildHookRegistry` signature from 3-arg to 4-arg

## Impact

- `internal/app/app.go` — saver wiring
- `internal/cli/doctor/checks/bootstrap_timing.go` — baseline exclusion
- `internal/cli/doctor/checks/runledger_workspace_isolation.go` — error propagation
- `internal/learning/suggestion.go` — drift counter time boundary
- `internal/runledger/workspace.go` — stash hint
