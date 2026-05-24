## Why

The cockpit Status page is intentionally read-only and refreshes itself on a timer, but the public cockpit docs currently focus on its sections and degraded states without saying how interaction works. That leaves operators without a clear cue that the page does not use cockpit help bindings and updates automatically.

## What Changes

- Update cockpit feature docs to describe the Status page as a read-only, auto-refreshing surface with no cockpit help-bar bindings.
- Update the downstream-docs-sync spec to require the same Status interaction coverage.

## Capabilities

### New Capabilities

### Modified Capabilities

- `downstream-docs-sync`: Public cockpit docs need Status interaction semantics that match the read-only runtime surface.

## Impact

- Affected docs/specs: `docs/features/cockpit.md`, `openspec/specs/downstream-docs-sync/spec.md`
- No runtime code changes
