## Context

The inline approval strip, the compact approval card, and the public docs all describe `s` as `allow session`, but the fullscreen diff dialog uses a shorter `session` label. This is a wording inconsistency rather than a behavior issue.

## Goals / Non-Goals

**Goals:**

- Make the fullscreen diff dialog use the same `allow session` wording as the rest of the chat approval surface.

**Non-Goals:**

- Change approval behavior.
- Redesign the approval dialog layout.

## Decisions

- Use the fuller `allow session` wording even in the diff dialog.
  - Rationale: consistency is more valuable here than the small width savings from the abbreviated label.

## Risks / Trade-offs

- [The action bar becomes slightly wider] → The help bar already truncates/clamps to available width and the wording gain is clearer than the cost.
