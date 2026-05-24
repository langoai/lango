## Context

The Tasks runtime and tests still cover navigation/help behaviors that used to exist in the main `tui-task-surface` spec, but repeated archives narrowed the merged requirement block. This is a main-spec maintenance problem rather than an implementation problem.

## Goals / Non-Goals

**Goals:**

- Rebuild the `Tasks page navigation` requirement block so the main spec again reflects the landed behavior.
- Preserve the currently valid scenarios that are already present in the main spec.

**Non-Goals:**

- Change runtime behavior.
- Rewrite unrelated task-surface requirement sections.

## Decisions

- Restore the affected navigation requirement as the union of currently landed scenarios rather than appending isolated one-off patches.
  - Rationale: the drift happened at the requirement-block level, so recovery should also operate at the requirement-block level.

- Keep the change spec-only.
  - Rationale: runtime/tests/docs already validate these behaviors; the missing artifact is the authoritative main spec text.

## Risks / Trade-offs

- [Recovered wording could diverge from runtime] → Only restore scenarios that are already evidenced by landed tests, docs, or archived spec deltas from today's work.
