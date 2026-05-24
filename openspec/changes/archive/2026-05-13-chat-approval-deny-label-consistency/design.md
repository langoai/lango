## Context

The runtime already treats both `d` and `Esc` as deny keys in approval state, but the visible labels differ between compact and fullscreen approval surfaces plus the approval-state help bar.

## Goals / Non-Goals

**Goals:**

- Make the deny label visually consistent wherever chat approval controls are shown.

**Non-Goals:**

- Change approval behavior.
- Redesign approval layout.

## Decisions

- Use `d/Esc` everywhere.
  - Rationale: it is compact, readable, and matches the actual key set without favoring one casing style over another.

## Risks / Trade-offs

- [Tiny wording churn across multiple surfaces] → The result is a more coherent approval UI with no behavioral risk.
