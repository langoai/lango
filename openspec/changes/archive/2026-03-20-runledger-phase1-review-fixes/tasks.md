## 1. Access Control (Fix 1)

- [x] 1.1 Remove orchestrator allowance from `checkRole` for `roleExecution` — return `ErrAccessDenied` when orchestrator calls execution-only tools
- [x] 1.2 Add test: `TestCheckRole_OrchestratorBlockedFromExecutionTools`
- [x] 1.3 Add test: `TestProposeStepResult_OrchestratorBlocked`

## 2. Approve Step Validation (Fix 2)

- [x] 2.1 Add validator type check in `buildRunApproveStep` — only `orchestrator_approval` steps allowed
- [x] 2.2 Add step status check — only `verify_pending` or `failed` status allowed
- [x] 2.3 Add test: `TestApproveStep_RejectNonOrchestratorApprovalType`
- [x] 2.4 Add test: `TestApproveStep_RejectWrongStatus`

## 3. WorkDir Phase-1 Readiness (Fix 3)

- [x] 3.1 Add `WorkDir` field to `ValidatorSpec` in `types.go`
- [x] 3.2 Add `cmd.Dir = spec.WorkDir` to `BuildPassValidator`, `TestPassValidator`, `FileChangedValidator`, `CommandPassValidator`
- [x] 3.3 Update `ArtifactExistsValidator` to use `filepath.Join(spec.WorkDir, target)` when WorkDir is set
- [x] 3.4 Add `PrepareStepWorkspace()` function to `workspace.go`
- [x] 3.5 Add `workspace *WorkspaceManager` field and `WithWorkspace()` method to `PEVEngine`
- [x] 3.6 Update `PEVEngine.Verify()` to call `PrepareStepWorkspace` when workspace is set
- [x] 3.7 Add test: `TestArtifactExistsValidator_WithWorkDir`
- [x] 3.8 Add test: `TestNeedsIsolation`
- [x] 3.9 Add test: `TestPEVEngine_WorkspaceIsolationFailure`

## 4. PEV Auto-Verification + Run Completion (Fix 4)

- [x] 4.1 Add `EventCriterionMet` event type and `CriterionMetPayload` to `journal.go`
- [x] 4.2 Add `EventCriterionMet` case to `applyEvent` in `snapshot.go`
- [x] 4.3 Add `AllStepsSuccessful()` method to `RunSnapshot` in `snapshot.go`
- [x] 4.4 Add `criterion_met` to Ent schema enum in `run_journal.go`
- [x] 4.5 Rewrite `buildRunProposeStepResult` to accept `*PEVEngine` and auto-trigger verification after journal append
- [x] 4.6 Extract `checkRunCompletion` shared function (AllStepsSuccessful → acceptance criteria → criterion_met journaling → completed/failed)
- [x] 4.7 Update `buildRunApproveStep` to use `checkRunCompletion`
- [x] 4.8 Update `BuildTools` to pass `pev` to `buildRunProposeStepResult`
- [x] 4.9 Add test: `TestProposeResult_AutoVerify_Pass`
- [x] 4.10 Add test: `TestProposeResult_AutoVerify_RunCompletion`
- [x] 4.11 Add test: `TestProposeResult_AutoVerify_CriteriaUnmet`
- [x] 4.12 Add test: `TestProposeResult_OrchestratorApproval_Flow`
- [x] 4.13 Add test: `TestProposeResult_InfraError`

## 5. Downstream Artifacts (Fix 5)

- [x] 5.1 Create `internal/cli/run/run.go` with `list`, `status`, `journal` subcommands
- [x] 5.2 Register run command in `cmd/lango/main.go` (GroupID: "auto")
- [x] 5.3 Add RunLedger to `collectFeatures` in `internal/cli/status/status.go`
- [x] 5.4 Update `status_test.go` with RunLedger enabled
- [x] 5.5 Update README.md: Features, CLI, Architecture, Configuration sections
- [x] 5.6 Create `docs/features/run-ledger.md` feature documentation
- [x] 5.7 Update `docs/features/index.md` with RunLedger card and status table row

## 6. Verification

- [x] 6.1 `go build ./...` passes
- [x] 6.2 `go test ./internal/runledger/` — all 48 tests pass
- [x] 6.3 `go test ./internal/cli/status/` — all tests pass
