## Context

The Tools page runtime distinguishes between a missing catalog (`Tool catalog is not available.`) and a configured catalog with zero categories (`No categories registered.`). The README shortcut table only mentions the first branch.

## Goals / Non-Goals

**Goals:**

- Make the README Tools row reflect both current degraded/empty-catalog branches.

**Non-Goals:**

- Change runtime behavior.
- Expand the README row into a full page walkthrough.

## Decisions

- Add the no-categories branch directly to the existing Tools description.
  - Rationale: the shortcut table is still an overview artifact, so one concise clause is enough.

## Risks / Trade-offs

- [The README row becomes slightly denser] → The added clause captures a real operator-visible state and remains compact.
