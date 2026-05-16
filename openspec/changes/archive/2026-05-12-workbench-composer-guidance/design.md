## Context

The workbench now has two distinct empty-state flows: setup recovery for incomplete profiles and starter prompts for ready profiles. The remaining inconsistency was the composer line itself, which still displayed generic copy regardless of the profile state.

## Goals / Non-Goals

**Goals:**
- Make the composer placeholder reflect the same state-aware guidance as the empty-state body.
- Preserve the behavior only for the standalone workbench surface.
- Keep the logic lightweight by reusing the existing readiness checks.

**Non-Goals:**
- Changing chat model global placeholder defaults.
- Adding dynamic prompt generation or clickable actions.
- Altering cockpit placeholder behavior.

## Decisions

### D1: Override placeholder in the page layer, not the chat model
The chat model still owns the global generic placeholder. The workbench-specific Mission Control page overrides the displayed placeholder only when the composer is empty and the page is rendering the workbench empty state.

### D2: Reuse the existing readiness heuristic
The placeholder split reuses the same incomplete-profile detection already used by the empty-state body so the two surfaces cannot drift apart easily.

## Risks / Trade-offs

- **[Two placeholder sources]** → The page now conditionally overrides the chat model placeholder. Mitigation: the override is narrowly scoped to the workbench empty state and leaves all other states untouched.
- **[More copy variance]** → Operators see different placeholders depending on readiness. Mitigation: that variance is intentional and improves correctness.
