## Why

Lango's workspace isolation feature (`runledger/workspace.go`) is fully implemented with worktree creation, patch export/apply, and config flag. However, the user-facing error messages are terse ("stash or commit before proceeding"), `git am` failure provides no recovery guidance, and `doctor` has no check for RunLedger workspace isolation health. The "Code Agent Orchestra" article emphasizes worktree isolation as a core enablement pattern — hardening the existing surface makes it production-ready.

## What Changes

- **Guided remediation in `CheckDirtyTree`**: Replace terse error with a summary of changed files and suggested commands (`git stash push -m "..."`)
- **Conflict guidance in `ApplyPatch`**: Wrap `git am` failure with explicit rollback instructions (`git am --abort`)
- **New doctor check `RunLedger Workspace Isolation`**: Config value, git availability, active worktree list via `git worktree list`, stale worktree detection by age threshold
- **Enablement conditions spec delta**: Document which validator types/modes should enable isolation by default
- **Doctor long description update**: Increment check count, add new check to list

## Capabilities

### New Capabilities
_(none — extends existing)_

### Modified Capabilities
- `run-ledger`: Delta spec adding enablement condition scenarios for workspace isolation

## Impact

- `internal/runledger/workspace.go` — `CheckDirtyTree` and `ApplyPatch` error messages
- `internal/cli/doctor/checks/` — new `runledger_workspace_isolation.go`, register in `AllChecks()`
- `internal/cli/doctor/doctor.go` — long description update
- `openspec/specs/run-ledger/` — delta spec
