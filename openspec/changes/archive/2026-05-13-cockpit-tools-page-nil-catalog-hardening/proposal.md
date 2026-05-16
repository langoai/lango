## Why

The cockpit sidebar always exposes a Tools destination conceptually, but the actual page is only registered when `application.ToolCatalog` is non-nil. That leaves a degraded UX gap: the docs say the Tools page shows an empty state when the catalog is unavailable, but the runtime instead omits the page registration path entirely.

The Tools page should fail closed into an explicit empty state, not disappear as a route.

## What Changes

- Make `pages.ToolsPage` nil-safe when the tool catalog is unavailable.
- Always register the Tools page in cockpit startup, even when the catalog is nil.
- Add regression coverage for the nil-catalog empty state.
- Sync the cockpit tools-page spec with the empty-state contract.

## Impact

- Cockpit navigation remains stable even in degraded or partial app wiring.
- The public cockpit docs about the Tools page empty state become true in runtime behavior.
