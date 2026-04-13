## Why

5 defects were found in RunLedger Phase 1 code review. PEV auto-execution not connected (steps permanently stall at verify_pending), approve without validator type verification, validators executing in main tree (workspace isolation meaningless), orchestrator allowed to call execution tools, downstream not reflected (CLI/TUI/README, etc.). These issues violate core invariant principles and require immediate fixes.

## What Changes

- Connect PEV auto-execution after journal recording in `run_propose_step_result` (propose → verify → completion check transitions in one call)
- Add validator type verification to `run_approve_step` (only orchestrator_approval type allowed, only verify_pending/failed states allowed)
- Add `WorkDir` field to `ValidatorSpec`, set `cmd.Dir` in all command-running validators
- Add `WorkspaceManager` field + `WithWorkspace()` method + `PrepareStepWorkspace()` call to `PEVEngine`
- Remove orchestrator permission for execution-only tools in `checkRole`
- Extract common `checkRunCompletion` function (AllStepsSuccessful → acceptance criteria → criterion_met journaling → completed/failed)
- Add new `EventCriterionMet` event type
- CLI: Add `lango run list|status|journal` subcommands
- Add RunLedger feature to `lango status` dashboard
- Update README, docs, openspec spec

## Capabilities

### New Capabilities

(none — all changes modify the existing run-ledger capability)

### Modified Capabilities

- `run-ledger`: PEV auto-verification, WorkDir injection, strict access control, run completion logic, EventCriterionMet, CLI downstream

## Impact

- `internal/runledger/tools.go` — PEV connection, checkRunCompletion, checkRole hardening
- `internal/runledger/pev.go` — workspace field + PrepareStepWorkspace call in Verify
- `internal/runledger/validators.go` — cmd.Dir = spec.WorkDir
- `internal/runledger/types.go` — ValidatorSpec.WorkDir
- `internal/runledger/workspace.go` — PrepareStepWorkspace function
- `internal/runledger/journal.go` — EventCriterionMet + CriterionMetPayload
- `internal/runledger/snapshot.go` — AllStepsSuccessful + applyEvent criterion_met
- `internal/ent/schema/run_journal.go` — criterion_met enum
- `internal/cli/run/run.go` — new CLI package
- `cmd/lango/main.go` — run command registration
- `internal/cli/status/status.go` — RunLedger feature
- README.md, docs/features/run-ledger.md, docs/features/index.md
