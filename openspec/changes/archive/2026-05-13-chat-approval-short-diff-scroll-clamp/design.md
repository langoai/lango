## Context

The approval dialog already computes the visible diff window from `contentHeight` and the total diff line count. The current upper clamp uses `len(lines)-1`, which is too loose when multiple lines fit on screen because it allows the window to start below the last meaningful page boundary.

## Goals / Non-Goals

**Goals:**

- Clamp the diff scroll offset to `max(len(lines)-visible, 0)` before slicing visible lines.
- Preserve existing scrolling behavior for genuinely long diffs.

**Non-Goals:**

- Change scroll step size.
- Redesign the approval dialog layout.

## Decisions

- Normalize `state.scrollOffset` inside `renderApprovalDialog()`.
  - Rationale: the function already owns the `visible` calculation and can clamp to the exact render-time limit.

## Risks / Trade-offs

- [Render path mutates approval state] → This is acceptable because the render path already updates diff-cache state, and the clamp is part of keeping the render state coherent.
