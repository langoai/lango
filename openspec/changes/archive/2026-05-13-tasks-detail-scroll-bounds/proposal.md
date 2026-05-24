## Why

The Tasks detail panel scroll offset currently increases without an upper bound and is only clamped opportunistically during render. That does not visibly break the page, but it lets state drift beyond the real scroll range and makes detail navigation less well-defined.

## What Changes

- Clamp task detail scrolling to the actual content height and visible detail height.
- Add regression coverage that `detailScroll` does not grow past the last meaningful offset.
- Update the task-surface spec to require bounded detail scrolling.

## Impact

- Detail scrolling behaves like a real bounded viewport instead of an unbounded counter.
- Runtime state, tests, and spec agree on the scroll contract.
