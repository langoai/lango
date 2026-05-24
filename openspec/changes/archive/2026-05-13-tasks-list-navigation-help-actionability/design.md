## Context

The Tasks page already supports list navigation, but the help surface does not distinguish between actionable multi-row task lists and inert zero/single-row states. Several other cockpit pages already hide inert navigation keys under the same condition.

## Goals / Non-Goals

**Goals:**

- Advertise Tasks list navigation only when another task row exists.
- Preserve detail-mode, cancel, and retry help behavior.

**Non-Goals:**

- Change task selection mechanics.
- Change detail-mode help contracts.

## Decisions

- Gate list-mode `↑/↓` help on `len(tasks) > 1`.
  - Rationale: that is the minimum condition for actual cursor movement to another row.

- Leave `Enter` visible in list mode even when no tasks exist.
  - Rationale: the page still uses `Enter` as its stable detail-toggle affordance, and this change is scoped only to inert navigation bindings.

## Risks / Trade-offs

- [Help count changes in empty/single-row states] → This is intentional because the goal is to show only actionable list-navigation keys.
