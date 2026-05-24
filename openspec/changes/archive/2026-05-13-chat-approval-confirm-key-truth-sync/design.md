## Context

`approvalState.confirmAction` already stores whether the pending confirmation is for `a` or `s`, and the key-handling path respects that state correctly. The mismatch is isolated to the render layer, which was previously given only a boolean confirm flag and therefore always displayed the `a` variant.

## Goals / Non-Goals

**Goals:**

- Make approval confirm prompts key-accurate across both inline and fullscreen approval surfaces.
- Preserve the existing confirmation logic and timing window.

**Non-Goals:**

- Change approval policy or safety classification.
- Redesign approval layout.

## Decisions

- Thread the full `approvalState` into the inline strip render helper and derive the current confirm key from that shared state.
- Derive the fullscreen dialog confirm key from the same `approvalState`.

These choices keep `approvalState.confirmAction` as the single source of truth for confirmation semantics.

## Risks / Trade-offs

- [Render helper signatures become slightly broader] → This is acceptable because it eliminates a concrete operator-facing lie and keeps the state model centralized.
