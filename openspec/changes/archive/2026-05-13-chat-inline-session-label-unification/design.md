## Context

The fullscreen approval dialog, approval banner/footer, status/help bar, and public docs all describe the session action as `allow session`, but the compact inline strip still uses a shorter `session` label. This is a wording inconsistency rather than a behavior issue.

## Goals / Non-Goals

**Goals:**

- Make the inline approval strip use the same `allow session` wording as the rest of the chat approval surface.

**Non-Goals:**

- Change approval behavior.
- Redesign the inline strip layout.

## Decisions

- Prefer consistency over the slightly shorter abbreviated label.
  - Rationale: the operator moves between multiple approval surfaces, so uniform language matters more than a small width saving.

## Risks / Trade-offs

- [The inline strip becomes a little wider] → The strip already truncates safely and the wording is more explicit.
