## Context

Current RunLedger code passes its package tests but still allows three important mismatches
between the intended Task OS model and runtime behavior:

1. `run_propose_step_result` journals a proposal before verifying that the target step
   exists, belongs to the caller, and is currently accepting proposals.
2. workspace isolation preparation reuses the same branch name for the same
   `runID/stepID`, so the second validation attempt fails after the first cleanup.
3. the app runtime still constructs `PEVEngine` without `WithWorkspace(...)`, while
   some user-facing docs imply isolation is already active.

This change fixes the first two in code and resolves the third by making the phase gate
explicit in runtime comments and docs, while leaving full activation to the planned
Phase 4 change.

## Goals

- Prevent unauthorized or malformed step proposals from polluting the journal.
- Make workspace preparation safe across retries and repeated validation attempts.
- Align code, docs, and OpenSpec on the fact that workspace isolation remains
  phase-gated and is not activated in Phase 1 runtime wiring.
- Leave a clean handoff to the planned Phase 2~4 changes.

## Non-Goals

- Enabling production workspace isolation in Phase 1.
- Implementing Ent-backed durable RunLedger persistence.
- Switching read paths to authoritative RunLedger snapshots.
- Integrating tool-profile enforcement into orchestrator runtime.

## Decisions

### 1. Validate step proposals before journaling

`run_propose_step_result` must become a guarded write path:

1. load the current snapshot
2. locate the target step
3. verify `step.OwnerAgent == caller`
4. verify the step is in an allowed pre-state (`in_progress`)
5. only then append `EventStepResultProposed`

This preserves the authority model: execution agents may only propose results for
their own active steps.

### 2. Keep workspace activation phase-gated

The cleanest resolution for the current review finding is not to silently enable
workspace isolation in Phase 1. Doing so would change runtime semantics beyond the
hardening scope and would bypass the staged rollout we already agreed on.

Instead:

- `PEVEngine.WithWorkspace(...)` remains available and tested
- `runLedgerModule` explicitly documents that Phase 1 does not call it
- README/docs/specs say the current release is readiness-only for isolation
- Phase 4 will activate it behind the dedicated rollout change

### 3. Make workspace lifecycle retry-safe

Two approaches were considered:

- delete the reused branch on cleanup
- generate a unique branch/path per validation attempt

We choose the second, plus best-effort branch cleanup:

- unique suffix avoids branch-exists races immediately
- cleanup still removes the worktree and deletes the generated branch
- repeated retries no longer depend on previous cleanup being perfect

### 4. Add direct tests for the failure modes

The review findings are subtle runtime bugs, not type-checking issues. The change must
include focused tests for:

- wrong agent proposing a step result
- proposing a result before the step is in progress
- repeated workspace preparation on the same step

## Risks / Trade-offs

- Keeping workspace disabled in Phase 1 means the review finding is resolved by
  explicit phase semantics rather than immediate activation. This is intentional and
  safer for the current rollout.
- Unique branch names create more short-lived refs, so cleanup must delete them
  proactively.
- Strict proposal preconditions may require a few existing tests to start emitting
  `step_started` events explicitly; this is desirable because it matches the intended
  state machine.
