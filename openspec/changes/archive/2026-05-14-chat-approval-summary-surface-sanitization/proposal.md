## Why

The inline approval strip already sanitizes malformed summaries, but the fallback approval banner and fullscreen approval dialog still render summary text raw. That leaves the approval surfaces inconsistent and allows malformed summaries to break the richer approval UIs.

## What Changes

- Sanitize approval summary text in the fallback banner and fullscreen dialog using the same plain single-line normalization as the inline strip.
- Add regression coverage for escaped and multiline summaries across those surfaces.
- Record the cross-surface summary contract in OpenSpec and downstream cockpit docs.

## Impact

- Unifies approval summary rendering across all chat approval surfaces.
- Prevents malformed tool summaries from destabilizing Tier 2 or fallback approval UIs.
