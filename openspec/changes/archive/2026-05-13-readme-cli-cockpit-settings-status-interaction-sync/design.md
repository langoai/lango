## Context

The runtime and cockpit feature docs already distinguish Settings from read-only pages by its embedded editor/footer, and already distinguish Status as a read-only auto-refresh surface. The README and CLI overview should reflect the same operator-facing reality.

## Goals / Non-Goals

**Goals:**

- Describe the Settings and Status interaction model in the public README and CLI overview.
- Keep the wording short enough for overview-level docs.

**Non-Goals:**

- Rewrite the full cockpit documentation.
- Add new runtime behavior.

## Decisions

- Update the README shortcut-table descriptions directly.
  - Rationale: that table is the fastest operator scan point for cockpit pages.

- Add one concise bullet to the `lango cockpit` CLI overview.
  - Rationale: the core CLI docs should set expectations without duplicating the full feature reference.

## Risks / Trade-offs

- [Overview docs become slightly denser] → Keep the added wording short and interaction-focused.
