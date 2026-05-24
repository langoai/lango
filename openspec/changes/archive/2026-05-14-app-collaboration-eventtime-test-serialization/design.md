## Context

The collaboration runtime bridge tests mostly run safely in parallel, but two of them temporarily replace the package-global `eventTime` seam. That pattern is the same class of flake already fixed in other packages: mutable global seam plus parallel execution.

## Goals / Non-Goals

**Goals:**
- Remove scheduler-dependent flakiness from the two seam-mutating tests
- Preserve the rest of the file's assertions and coverage

**Non-Goals:**
- Refactoring production collaboration runtime code
- Reworking the `eventTime` seam itself
- Changing any event payload semantics

## Decisions

Remove `t.Parallel()` only from the tests that override `eventTime`.
Rationale: it is the smallest safe fix and preserves parallelism elsewhere in the file.

## Risks / Trade-offs

- [Trade-off] Two tests lose parallel speed. → Mitigation: this is negligible compared with deterministic suite behavior.
