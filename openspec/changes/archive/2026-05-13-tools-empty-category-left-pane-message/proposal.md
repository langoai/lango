## Why

When the Tools page has a configured catalog but zero registered categories, the right panel explains `No categories registered.` but the left category pane is mostly blank. That makes the empty catalog state harder to interpret than it needs to be.

## What Changes

- Render an explicit `No categories registered.` message in the left category pane as well.
- Add regression coverage for the empty-category catalog state.
- Update the Tools page spec and cockpit feature docs to describe the clearer empty-catalog presentation.

## Impact

- The empty category-browser state becomes self-explanatory from either pane.
- Runtime, tests, docs, and spec all align on the same empty-category contract.
