## Context

The Tasks page already computes `detailMaxScroll(task)` to bound scrolling, so the runtime knows when detail content has no overflow. The mismatch is only in the help surface, which still hard-codes scroll bindings for every detail-mode state.

## Goals / Non-Goals

**Goals:**

- Make Tasks detail-mode help advertise scroll keys only when scroll is possible.
- Preserve the existing detail toggle, close, cancel, and retry help behavior.

**Non-Goals:**

- Change detail scrolling mechanics.
- Redesign the Tasks detail layout.

## Decisions

- Use `detailMaxScroll(*selectedTask) > 0` as the gate for scroll help.
  - Rationale: this is the same overflow signal the runtime already uses for actual scrolling bounds.

- Keep `Enter` and `Esc` visible in detail mode even when scrolling is unavailable.
  - Rationale: those actions remain valid regardless of overflow.

## Risks / Trade-offs

- [Detail-mode help length changes based on content size] → This is intentional because the goal is to show only actionable keys.
