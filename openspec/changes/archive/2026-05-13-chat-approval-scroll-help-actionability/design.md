## Context

The fullscreen diff approval dialog already computes the visible diff slice from `contentHeight`, the fixed dialog chrome, and `scrollOffset`. The missing piece is that the action bar does not consult the same overflow condition before advertising scroll keys.

## Goals / Non-Goals

**Goals:**

- Hide approval-dialog scroll help when the diff preview has no overflow.
- Preserve scroll help for genuinely scrollable diffs.

**Non-Goals:**

- Change diff rendering itself.
- Change scroll step size or split-mode behavior.

## Decisions

- Derive a `hasScrollableDiff` condition from `len(allLines) > visible`.
  - Rationale: this matches the dialog's actual visible slice calculation.

## Risks / Trade-offs

- [Help changes with dialog height] → This is intentional because scrollability itself changes with available height.
