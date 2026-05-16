## Context

The approval surfaces already render `d/Esc denies` during confirm-pending state, so the docs should explain that the operator still has an immediate deny exit instead of describing only the confirm path.

## Goals / Non-Goals

**Goals:**

- Surface the confirm-pending deny path in public cockpit docs.

**Non-Goals:**

- Change approval behavior.
- Expand the docs into a full approval state machine.

## Decisions

- Add one explicit line to the existing double-press guardrail list rather than creating a separate subsection.
  - Rationale: the behavior belongs directly beside the confirm steps it modifies.

## Risks / Trade-offs

- [Docs become slightly more detailed] → The extra sentence covers a real operator-visible escape path and remains concise.
