## Why

Mission Control already hides `↑/↓` in the true empty state, but once any content exists it always advertises navigation keys even when the currently focused lane has no alternative row to move to. That leaves inert help in single-row or single-item situations.

## What Changes

- Show Mission Control `↑/↓` help only when the currently focused lane actually has another row to move to.
- Add regressions for single-row and multi-row help states.
- Update cockpit-pages spec and feature docs to describe the stricter navigation-help contract.

## Impact

- Mission Control stops advertising inert navigation in single-row states.
- Help becomes more accurately tied to the currently focused lane.
