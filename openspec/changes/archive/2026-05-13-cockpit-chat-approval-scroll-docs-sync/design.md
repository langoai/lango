## Context

The chat approval dialog now shows scroll controls only when the diff preview actually overflows, but the feature doc still summarizes Tier 2 approval as an overlay "with ... scroll" without the overflow condition.

## Goals / Non-Goals

**Goals:**

- Make the cockpit feature docs describe Tier 2 approval scrolling accurately.

**Non-Goals:**

- Change runtime behavior.
- Expand the Chat section into a full approval-dialog manual.

## Decisions

- Add the overflow condition directly to the Tier 2 bullet.
  - Rationale: that is the shortest accurate wording and keeps the section compact.

## Risks / Trade-offs

- [Docs become slightly more conditional] → This is worthwhile because the new runtime behavior is specifically about suppressing inert scroll affordances.
