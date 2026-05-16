## Context

The exec tool tests are mostly parallel-safe, but `TestWarnFallbackOnce_WritesOnlyOneWarning` temporarily replaces the package-global `execWarningWriter`. That pattern is similar to the workflow seam race already fixed earlier: a mutable global plus parallel execution is unnecessary risk.

## Goals / Non-Goals

**Goals:**
- Remove the parallelism from the single test that mutates `execWarningWriter`
- Preserve the test's current assertions and scope

**Non-Goals:**
- Refactoring runtime exec code
- Reworking the warning seam itself
- Changing any user-facing warning behavior

## Decisions

Remove `t.Parallel()` from the warning seam test.
Rationale: this is the smallest safe fix and keeps the rest of the file parallel where appropriate.

## Risks / Trade-offs

- [Trade-off] One test loses parallel speed. → Mitigation: the cost is negligible compared with deterministic suite behavior.
