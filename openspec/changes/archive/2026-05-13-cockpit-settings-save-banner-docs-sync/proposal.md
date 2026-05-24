## Why

The embedded Settings editor already shows inline save feedback at the top of the menu after an embedded save attempt, but the public cockpit docs currently mention only persistence availability and key discoverability. That leaves operators without a documented expectation for where successful or failed save feedback appears.

## What Changes

- Update cockpit feature docs to describe the inline save-success and save-failure banners on the Settings page.
- Update the downstream-docs-sync spec to require the same Settings save-feedback coverage.

## Capabilities

### New Capabilities

### Modified Capabilities

- `downstream-docs-sync`: Public cockpit docs need Settings save-feedback coverage that matches the embedded editor.

## Impact

- Affected docs/specs: `docs/features/cockpit.md`, `openspec/specs/downstream-docs-sync/spec.md`
- No runtime code changes
