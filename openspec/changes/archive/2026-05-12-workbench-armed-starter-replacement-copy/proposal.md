## Why

The workbench already lets `1/2/3` replace an armed starter prompt, but that capability was not clearly reflected in the public copy. The UI should advertise the actual next actions available to the operator.

## What Changes

- Update ready-profile workbench docs to state that `1/2/3` remain available after a starter is armed.
- Keep the existing submit-focused seeded-state guidance intact.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `downstream-docs-sync`: Public workbench docs now explain that armed starter prompts can still be replaced with `1/2/3`.

## Impact

- Affected docs: `README.md`, `docs/cli/core.md`, `docs/features/cockpit.md`
