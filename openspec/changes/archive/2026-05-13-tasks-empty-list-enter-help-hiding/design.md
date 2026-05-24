## Context

The Tasks runtime already fails closed on `Enter` when `len(tasks) == 0`, but the list-mode help surface still unconditionally shows `Enter: details`. This is a help/runtime mismatch, not a data or layout problem.

## Goals / Non-Goals

**Goals:**

- Hide inert `Enter` help in the zero-task list state.
- Keep `Enter` help for any list state with a selected task row.

**Non-Goals:**

- Change detail-mode behavior.
- Change list-mode navigation help beyond the existing separate actionability work.

## Decisions

- Gate list-mode `Enter` help on `len(tasks) > 0`.
  - Rationale: detail toggling requires a selected task row, and the runtime already uses the same condition.

## Risks / Trade-offs

- [Empty-state help becomes fully blank] → This is acceptable because no page-local key remains actionable in the zero-task list state.
