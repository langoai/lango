## Context

The current runtime and tests still cover several cockpit behaviors that used to exist in the main `cockpit-pages` spec, but repeated archives on the same capability have narrowed the merged requirement blocks. This is a spec-maintenance problem rather than an implementation problem.

## Goals / Non-Goals

**Goals:**

- Rebuild the affected `cockpit-pages` requirement blocks so the main spec again reflects the landed cockpit behavior.
- Preserve all currently valid scenarios that are already present in the main spec.

**Non-Goals:**

- Change runtime behavior.
- Rewrite unrelated cockpit capability sections.

## Decisions

- Restore the affected requirement blocks with the union of currently landed scenarios rather than adding one-off patch scenarios.
  - Rationale: the merge drift happened at the requirement-block level, so recovery should also operate at the requirement-block level.

- Keep the change spec-only.
  - Rationale: runtime/tests already validate these behaviors; the missing artifact is the authoritative main spec text.

## Risks / Trade-offs

- [Recovered wording could diverge from current runtime] → Only restore scenarios that are already evidenced by landed tests, docs, or archived spec deltas from today’s work.
