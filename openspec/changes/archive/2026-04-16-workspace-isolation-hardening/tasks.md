## 1. Guided remediation in CheckDirtyTree

- [x] 1.1 In `internal/runledger/workspace.go`, enhance `CheckDirtyTree` error: count changed files, suggest `git stash push -m "lango-workspace-isolation"`
- [x] 1.2 Add unit test for the guided error message format

## 2. Conflict guidance in ApplyPatch

- [x] 2.1 In `internal/runledger/workspace.go`, wrap `ApplyPatch` failure with rollback instructions: include `git am --abort` in error message
- [x] 2.2 Add unit test verifying the error message contains rollback guidance

## 3. RunLedger Workspace Isolation doctor check

- [x] 3.1 Create `internal/cli/doctor/checks/runledger_workspace_isolation.go` — check config value, git availability, parse `git worktree list --porcelain` for active/stale worktrees
- [x] 3.2 Register in `checks.go` `AllChecks()`
- [x] 3.3 Update `doctor.go` long description: add "RunLedger Workspace Isolation" to Execution category, increment count 26→27
- [x] 3.4 Add unit tests: isolation disabled→skip, nil config→skip, enabled→not-fail, name check

## 4. OpenSpec delta spec

- [x] 4.1 Delta spec covers guided remediation, conflict guidance, enablement conditions, doctor check scenarios

## 5. Verification

- [x] 5.1 `go build ./...` passes
- [x] 5.2 `go test ./...` passes — all modified packages green, zero FAIL
