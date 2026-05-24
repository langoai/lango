## Why

The cockpit Settings page already embeds a rich settings editor with its own inline help footer, but the public cockpit docs currently only describe persistence behavior and do not tell operators how key discoverability works on that page. That leaves the docs underselling a real part of the shipped UI surface.

## What Changes

- Update cockpit feature docs to describe that the Settings page uses the embedded editor's own inline help/footer instead of the cockpit help bar.
- Describe the primary menu/form key surface at a high level using the actual embedded editor behavior.
- Update the downstream-docs-sync spec to require this Settings help-surface coverage.

## Capabilities

### New Capabilities

### Modified Capabilities

- `downstream-docs-sync`: Public cockpit docs need a Settings key-surface description that matches the embedded editor.

## Impact

- Affected docs/specs: `docs/features/cockpit.md`, `openspec/specs/downstream-docs-sync/spec.md`
- No runtime code changes
