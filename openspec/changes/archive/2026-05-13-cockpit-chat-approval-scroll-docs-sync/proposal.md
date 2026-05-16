## Why

The fullscreen approval dialog now hides `↑/↓ scroll` help when the diff preview fully fits, but the public cockpit feature docs still describe Tier 2 approval as if scrolling were always part of that surface. That leaves the docs slightly broader than the actual runtime contract.

## What Changes

- Update cockpit feature docs to describe Tier 2 approval scrolling as conditional on diff overflow.
- Update the downstream-docs-sync spec to require the same Chat approval-scroll wording.

## Capabilities

### New Capabilities

### Modified Capabilities

- `downstream-docs-sync`: Public cockpit docs need Tier 2 approval-scroll wording that matches the current runtime contract.

## Impact

- Affected docs/specs: `docs/features/cockpit.md`, `openspec/specs/downstream-docs-sync/spec.md`
- No runtime code changes
