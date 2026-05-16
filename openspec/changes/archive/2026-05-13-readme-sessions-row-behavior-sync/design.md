## Context

The Sessions page runtime and feature docs now distinguish unavailable vs configured-empty history and enforce newest-first ordering. The README shortcut table is the remaining lagging overview artifact.

## Goals / Non-Goals

**Goals:**

- Surface the current Sessions behavior contract in the README shortcut table.

**Non-Goals:**

- Change runtime behavior.
- Duplicate the full Sessions feature-doc section.

## Decisions

- Add a short newest-first plus empty/unavailable clause directly to the README row.
  - Rationale: this keeps the shortcut table compact while still conveying the important behavior split.

## Risks / Trade-offs

- [The README row becomes a bit denser] → The added wording captures real operator-facing behavior and remains readable.
