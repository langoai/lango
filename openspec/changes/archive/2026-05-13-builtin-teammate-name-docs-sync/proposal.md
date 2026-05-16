## Why

Several public docs still used older built-in teammate names like `executor` and `researcher` in workflow and preset examples, even though the current built-in registry is `operator`, `navigator`, `vault`, `librarian`, `automator`, `planner`, `chronicler`, and `ontologist`. That leaves user-facing examples out of sync with the current runtime.

## What Changes

- Update the README workflow example to use `operator` and `librarian`.
- Update the CLI workflow status example to use `operator`, `librarian`, and `planner`.
- Update the config preset feature page to reference current built-in teammate names.
- Extend downstream docs sync coverage for current built-in teammate names in public examples.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `downstream-docs-sync`: public examples now use the current built-in teammate names.

## Impact

- Affected docs: `README.md`, `docs/cli/automation.md`, `docs/features/config-presets.md`
- Affected specs: `openspec/specs/downstream-docs-sync/spec.md`
