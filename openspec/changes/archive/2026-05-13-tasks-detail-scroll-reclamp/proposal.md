## Why

Task detail scrolling is now bounded while the operator presses `↓`, but `detailScroll` can still remain stale when the selected task content shrinks after a refresh. In that case the render path only clamps to `len(lines)-1`, which can still land below the last meaningful viewport offset and show an overly truncated detail panel.

## What Changes

- Re-clamp `detailScroll` to `detailMaxScroll(...)` when rendering or refreshing task detail.
- Add regression coverage for content shrink after scrolling.
- Update the task-surface spec to require detail scroll offsets to stay within the effective viewport range after refresh.

## Impact

- Detail panels stay stable even when task content changes while open.
- Scroll state becomes a true bounded viewport offset rather than a stale counter.
