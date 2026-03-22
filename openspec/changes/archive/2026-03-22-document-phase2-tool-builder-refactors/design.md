## Context

Four recent commits finished a second wave of tool-builder extraction and follow-up verification:

- `4f59b8c`: package-owned builders for agent memory, automation, data, team, and sentinel tools
- `eddbe4b`: package-owned builders for foundation tools plus additional wiring tests
- `c3a9322`: parity coverage for extracted builders
- `bd97b6c`: agent memory recall + kind validation bugfixes

Existing main specs already cover some adjacent behavior, but they lag behind the current ownership boundaries and the new agent memory validation semantics.

## Goals / Non-Goals

**Goals:**
- Capture the stable contracts introduced by the recent commits.
- Place each change under an existing capability when possible instead of inventing unnecessary new specs.
- Keep the documented architecture consistent with current Core/Application boundaries.

**Non-Goals:**
- Re-document unchanged user-facing behavior for every moved tool.
- Re-open broader architecture topics already deferred to future refactoring phases.
- Normalize unrelated historical mismatches outside the recent-commit scope.

## Decisions

### Decision: Update existing capabilities instead of creating a new umbrella spec

The recent commits span multiple parts of the system, but their effects fit naturally into existing capabilities:

- `agent-memory` for package-owned memory tools and validation semantics
- `automation-agent-tools` for shared automation contracts
- `domain-tool-builders` for package ownership of builders
- `tool-catalog` for registration expectations
- `parity-verification` for regression coverage

This keeps the main spec set discoverable and avoids a one-off “refactoring bucket” capability.

### Decision: Document ownership boundaries only where they are now stable

The refactors intentionally left some builders in `internal/app/` because of real dependency constraints:

- `buildMetaTools` still spans `knowledge` and `learning`
- `buildOnChainEscrowTools` still depends on `escrow`/`hub` type knowledge

The synced specs should document these as explicit exceptions instead of pretending every builder has moved.

### Decision: Treat agent memory validation as a behavior change, not just an implementation detail

The agent memory bugfix changed observable behavior:

- invalid `kind` values now fail instead of being silently accepted
- `memory_agent_recall` preserves context fallback even when `kind` filtering is used

These belong in the main contract because they change what callers can rely on.

## Risks / Trade-offs

- [Risk] Spec overlap between `domain-tool-builders` and `tool-catalog`
  - Mitigation: put builder ownership in `domain-tool-builders`, and keep `tool-catalog` focused on registration outcomes.

- [Risk] Over-specifying internal helper details
  - Mitigation: document shared interfaces and helper behavior only where they are now cross-package contracts.

- [Risk] Leaving older unrelated mismatches untouched
  - Mitigation: scope this change to the recent four commits and avoid mixing in unrelated cleanup.

## Migration Plan

1. Create a documentation-only change describing the recent refactoring commits.
2. Add delta specs for the affected existing capabilities.
3. Sync the delta specs into `openspec/specs/`.
4. Archive the documentation change once main specs reflect the current code.

## Open Questions

- None for this documentation pass.
