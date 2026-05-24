## Context

The mission-control projector tests are mostly parallel-safe, but several of them temporarily replace the per-projector `nowFn` seam to freeze time-sensitive summaries and trend calculations. Those tests should not run in parallel with sibling tests that could observe the same seam unexpectedly.

## Goals / Non-Goals

**Goals:**
- Remove scheduler-dependent flakiness from the projector tests that override `nowFn`
- Preserve all existing assertions and coverage

**Non-Goals:**
- Refactoring production mission-control projector code
- Reworking the time seam itself
- Changing any user-visible snapshot behavior

## Decisions

Remove `t.Parallel()` only from the tests that override `projector.nowFn`.
Rationale: it is the smallest safe fix and keeps parallelism elsewhere in the file.

## Risks / Trade-offs

- [Trade-off] A handful of tests lose parallel speed. → Mitigation: deterministic suite behavior is more valuable than minimal local speedup.
