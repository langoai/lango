## Context

The chat runtime and tests still cover a broad approval/help surface: idle versus streaming quit semantics, approval key wording, fullscreen diff affordances, and confirm-pending guidance. The main `tui-chat-rendering` spec no longer reflects most of those landed contracts after repeated archives on the same requirement block.

## Goals / Non-Goals

**Goals:**

- Rebuild the `Turn state strip` requirement block so the main spec reflects the landed chat behavior.
- Preserve any currently valid scenarios already present in the main spec.

**Non-Goals:**

- Change chat runtime behavior.
- Rewrite unrelated `tui-chat-rendering` sections.

## Decisions

- Restore the affected requirement block as the union of currently landed scenarios from archived deltas.
  - Rationale: the drift happened at the requirement-block level, so recovery should operate at the same granularity.

- Keep the change spec-only.
  - Rationale: runtime/tests/docs already validate these behaviors; the missing artifact is the authoritative main spec text.

## Risks / Trade-offs

- [Recovered wording could diverge from runtime] → Only restore scenarios already evidenced by landed tests, docs, or archived spec deltas from today's chat-surface work.
