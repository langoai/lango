## Why

The runtime now keeps `d/Esc` denial visible even while a critical-risk approval is in confirm-pending state, but the public cockpit feature docs still explain only the repeat-confirm path and the generic "different key resets" rule. That leaves the immediate deny escape path under-documented.

## What Changes

- Update cockpit feature docs to mention that `d` or `Esc` still deny immediately while confirm-pending is active.
- Update the downstream-docs-sync spec to require the same approval-confirmation wording.

## Capabilities

### New Capabilities

### Modified Capabilities

- `downstream-docs-sync`: Public cockpit docs need confirm-pending deny-path coverage for chat approvals.

## Impact

- Affected docs/specs: `docs/features/cockpit.md`, `openspec/specs/downstream-docs-sync/spec.md`
- No runtime code changes
