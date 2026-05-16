## Why

The cockpit Tools page exposes `Esc back` in its help bindings even though the page does not implement any `Esc` behavior. That is misleading operator guidance in a high-visibility surface.

## What Changes

- Remove the stale `Esc back` help binding from the Tools page.
- Update the cockpit tools-page spec to match the actual up/down-only page navigation contract.

## Impact

- The Tools page help bar no longer advertises a key that does nothing.
- The tools-page contract matches what the page really supports.
