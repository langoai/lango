## Context

Workspace preparation and WorkDir-aware validators exist before Phase 4, but the runtime still
executes without active isolation. Tool access is also broader than a production Task OS should
allow. Phase 4 turns the prepared pieces on and narrows the execution surface.

## Goals

- Enable workspace isolation in the real app runtime.
- Ensure repeated isolated validations are retry-safe.
- Enforce tool-profile narrowing per step.
- Keep failure semantics fail-closed and visible to the orchestrator.

## Non-Goals

- Changing planner schema or write-through ID semantics.
- Replacing the PEV/policy model introduced earlier.

## Decisions

### 1. Isolation is opt-in by explicit config gate

The runtime enables `WithWorkspace(NewWorkspaceManager())` only when
`runLedger.workspaceIsolation` is enabled, not merely when RunLedger exists.

### 2. Tool governance is step-scoped

Execution agents see only the tool profile attached to the active step. Orchestrator/supervisor
paths keep their own narrow profile.

### 3. Retry-safe workspace lifecycle is mandatory

Repeated validation of the same step must not fail due to reused branch or worktree identities.
