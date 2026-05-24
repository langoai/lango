## Why

The fullscreen approval dialog renderer assumes that approval state is always preinitialized. In current call paths that is usually true, but the renderer itself dereferences `state.diffCache` whenever diff content exists. That makes the renderer fragile: a nil state can panic instead of failing closed.

## What Changes

- Make the fullscreen approval dialog renderer tolerate a nil approval state.
- Add a regression that renders diff content with a nil state and verifies it does not panic.
- Record the renderer-hardening contract in OpenSpec.

## Impact

- Improves chat approval renderer robustness without changing visible operator behavior.
- Reduces the chance that a future caller or test harness can crash the TUI through an uninitialized approval state.
