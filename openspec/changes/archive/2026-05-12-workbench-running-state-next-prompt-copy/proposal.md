## Why

The workbench already supports staging a next prompt while the current turn is streaming, but the running-state copy still undersold that capability. The UX should teach the highest-leverage next action available in that state.

## What Changes

- Update the running-state body, placeholder, and footer copy to mention typing the next prompt and pressing `Enter` to interrupt-and-run it.
- Update docs to match that behavior.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `mission-workbench-tui`: Running-state guidance now advertises next-prompt staging and interrupt-and-run behavior.
- `downstream-docs-sync`: Public docs now describe the running-state next-prompt loop.

## Impact

- Affected code: `internal/cli/cockpit/pages/missioncontrol_workbench.go`
- Affected tests: `internal/cli/workbench/model_test.go`
- Affected docs: `README.md`, `docs/cli/core.md`, `docs/features/cockpit.md`
