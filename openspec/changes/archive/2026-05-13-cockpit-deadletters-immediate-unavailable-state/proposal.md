## Why

The Dead Letters page now stays registered even when its bridge callbacks are absent, but before the first activation-load cycle it can still render the generic "No current dead-letter backlog." empty state. That is misleading: the subsystem is unavailable, not simply empty.

The page should surface unavailable messaging immediately, not only after an activation-triggered load error arrives.

## What Changes

- Make the Dead Letters page render unavailable messaging immediately when no backlog list callback is configured.
- Add a regression for the immediate unavailable-state view.
- Sync the cockpit page spec to require the direct unavailable render path.

## Impact

- Dead Letters matches the degraded-state quality bar already used by Tasks and Approvals.
- Operators see the correct state before any background load cycle runs.
