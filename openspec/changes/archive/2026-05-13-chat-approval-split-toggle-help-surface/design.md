## Context

`handleApprovalDialogKey()` already handles `t` by toggling split mode and invalidating the diff cache. The missing piece is pure discoverability: operators only see the mode reflected after they somehow know to press `t`.

## Goals / Non-Goals

**Goals:**

- Surface the split-toggle key anywhere the fullscreen approval dialog already shows other diff-related controls.

**Non-Goals:**

- Change split-mode behavior.
- Add split-mode support to the inline approval strip.

## Decisions

- Show `t split` only when diff content exists.
  - Rationale: the toggle is meaningless without a diff preview.

## Risks / Trade-offs

- [Action bar becomes slightly denser] → The added entry is a real existing capability and only appears in the richer diff-dialog surface.
