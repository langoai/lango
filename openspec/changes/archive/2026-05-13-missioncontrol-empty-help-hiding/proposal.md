## Why

Mission Control already renders a distinct empty state when there are no missions, no pending decision, and no loops, but its help bar still advertises `↑/↓` navigation even though those keys are inert in that state.

## What Changes

- Hide Mission Control `↑/↓` help in the true empty state.
- Add regressions for empty-state and populated-state help.
- Update cockpit page specs and feature docs to describe the reduced empty-state help surface.

## Impact

- The Mission Control empty state stops advertising inert navigation keys.
- Runtime help, tests, docs, and spec describe the same empty-state contract.
