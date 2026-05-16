## Why

The chat runtime accepts `Ctrl+D` as an immediate quit path in every turn state, including `approving`. The approving-state help bar currently shows only `a`, `s`, and `d/Esc`, so it hides a still-actionable global quit key.

## What Changes

- Add `Ctrl+D quit` to the approving-state help bar.
- Add a regression for the approving help surface.
- Sync docs/spec wording with the approving-state quit contract.

## Impact

- Removes a real key-discoverability mismatch from the chat approval state.
- Keeps approval-state help aligned with the runtime's global quit path.
