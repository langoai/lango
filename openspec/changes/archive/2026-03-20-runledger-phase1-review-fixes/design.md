## Context

RunLedger Phase 1 was implemented as a scaffold, but 5 defects were found during code review:

1. PEV auto-execution is not connected to `run_propose_step_result`, causing steps to permanently stall at `verify_pending`
2. `run_approve_step` passes any step without validator type verification
3. Validators execute in the main tree — workspace isolation is meaningless without WorkDir support
4. `checkRole` also allows orchestrator in execution tools — access control is undermined
5. Downstream artifacts not reflected (CLI/TUI/README, etc.)

The code implementation is already complete, and this design document serves as a post-hoc record.

## Goals / Non-Goals

**Goals:**
- PEV auto-verification: connect propose → journal → verify → completion check in a single call
- Strict access control: enforce strict orchestrator vs. execution agent role separation
- Validator WorkDir: prepare field/support in Phase 1, activate with one line in Phase 3
- Run completion: automatic acceptance criteria check after step verification → run state transition
- Downstream: reflect in CLI, status dashboard, README, docs, openspec

**Non-Goals:**
- Actual worktree activation (Phase 3)
- DB transaction wrapping (Phase 2 — when switching to Ent store)
- TUI-specific RunLedger surface (unnecessary since it's a runtime feature, not config-driven)

## Decisions

### 1. PEV auto-execution location: inside `buildRunProposeStepResult`

Call `pev.Verify()` immediately after journal recording in the `run_propose_step_result` handler. Synchronous call chosen over separate event handler or async processing because:
- Phase 1 uses MemoryStore (sequential calls) — synchronous is simplest
- PEV results must be immediately included in tool responses so the agent can decide next actions
- When switching to Ent store in Phase 2, only `tx := client.Tx()` wrapping needs to be added

Alternative: Async trigger via event bus → increased complexity, unnecessary in Phase 1

### 2. Dual error handling: infrastructure vs. business

- Infrastructure failure (validator not registered, exec failure): `return nil, fmt.Errorf(...)` — non-nil Go error
- Business failure (validation not passed): structured map payload, nil error

This distinction is needed because: when an agent receives a Go error, it treats the tool call itself as failed; when it receives a structured payload, it can make policy decisions within the normal flow.

### 3. orchestrator_approval flow: PEV auto-execution → always failed → approve

The orchestrator_approval validator always returns failed. When PEV auto-executes, the step transitions to `failed`. Therefore, `run_approve_step` must allow not only `verify_pending` but also `failed` state.

Alternative: PEV detects orchestrator_approval and skips → step stays at verify_pending until approve → special case logic leaks into PEV, so rejected

### 4. Extract common checkRunCompletion function

Both `run_propose_step_result` and `run_approve_step` need to check run completion after step completion. Same logic:
- `AllStepsSuccessful()` → acceptance criteria verification → `EventCriterionMet` journaling → completed/failed
- `AllStepsTerminal()` but not successful → run failed
- In progress → running

### 5. WorkDir: field added + validator support, activation in Phase 3

Add `ValidatorSpec.WorkDir` field and use it in all command-running validators. In Phase 1, WorkDir is always an empty string (maintaining existing behavior). Activate in Phase 3 with one line: `pev.WithWorkspace(NewWorkspaceManager())`.

## Risks / Trade-offs

- [PEV synchronous call performance] → Negligible in Phase 1 with in-memory store. Address with validator timeout settings in Phase 2.
- [checkRunCompletion duplicate criterion_met journaling] → Already met criteria are recorded in journal every time. Add "skip if already met" condition in Phase 2.
- [orchestrator_approval 2-step transition (verify_pending → failed → approved)] → validation_failed event is recorded in journal followed by validation_passed. Final state is correct on replay but event log is somewhat verbose. Acceptable trade-off.
