## Context

The approval surfaces now correctly render either `Press 'a' again...` or `Press 's' again...` depending on the pending confirm action. The remaining mismatch is only in the prose description inside the cockpit feature reference.

## Goals / Non-Goals

**Goals:**

- Document the double-press guardrail in a way that matches the runtime prompt semantics.

**Non-Goals:**

- Change approval behavior.
- Rewrite the entire approval section.

## Decisions

- Replace the hard-coded `a` example with wording that explicitly refers to repeating the same pending action key.
  - Rationale: this matches both the inline and fullscreen approval surfaces without adding much complexity.

## Risks / Trade-offs

- [Docs become a touch less concrete] → The benefit is that the text becomes correct for both approval actions rather than only one.
