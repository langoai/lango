## Why

RunLedger remediation work completed the runtime, CLI, and documentation surfaces, but the product surfaces are still inconsistent: `lango settings` cannot edit RunLedger config, and `lango doctor` does not validate RunLedger-specific configuration invariants. This leaves one of the system's major subsystems configurable only through raw config editing and harder to diagnose operationally.

## What Changes

- Add a dedicated `RunLedger` category to the `lango settings` TUI and expose all `config.RunLedger` fields as editable form inputs.
- Wire the new settings form into the hierarchical menu, setup flow, and `ConfigState.UpdateConfigFromForm`.
- Add a `RunLedgerCheck` to `lango doctor` that validates enabled-state invariants and key value sanity.
- Update doctor help text and settings help text so RunLedger appears in the documented coverage.
- Update README and affected docs to reflect that RunLedger now has a settings surface and a doctor check.

## Capabilities

### New Capabilities

_(none)_

### Modified Capabilities

- `cli-settings`: Add RunLedger configuration coverage to the settings TUI
- `cli-doctor`: Add RunLedger-specific diagnostics to the doctor command

## Impact

- **Code**: `internal/cli/settings/*`, `internal/cli/tuicore/state_update.go`, `internal/cli/doctor/checks/*`, `internal/cli/doctor/doctor.go`
- **Docs**: `README.md`, `docs/cli/*`, any settings/doctor text that enumerates configuration coverage
- **Risk**: Low — additive UI and diagnostics surface; no core runtime behavior changes
