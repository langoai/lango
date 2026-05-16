## Context

`handleKey()` routes both `stateIdle` and `stateFailed` into `handleIdleKey()`, so the quit path is already shared. The mismatch is only in the `/help` wording and any public docs that summarize the same key contract.

## Goals / Non-Goals

**Goals:**

- Make `/help` describe the real idle/failed `Ctrl+C` semantics.
- Keep the change limited to wording and regression coverage.

**Non-Goals:**

- Change actual key behavior.
- Change streaming-state cancellation wording.

## Decisions

- Update `/help` wording instead of changing runtime logic.
  - Rationale: the implementation is already correct; the user-facing copy is the lagging artifact.

## Risks / Trade-offs

- [The help text becomes slightly longer] → The added `failed` qualifier is small and removes ambiguity.
