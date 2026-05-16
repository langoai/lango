## Why

The standalone `lango` workbench is the default interactive entry point, but its empty Mission Control state currently tells the operator to start chatting even when the active profile is obviously incomplete. That creates a dead-end first-run experience instead of guiding the operator toward setup and verification.

## What Changes

- Detect incomplete workbench configuration at the Mission Control empty state.
- Show direct setup guidance to `lango onboard`, `lango settings`, and `lango doctor` when the active profile is incomplete.
- Keep the extra setup guidance out of the empty state once the profile has a usable provider, model, and authentication path.
- Document the new workbench guidance in README and CLI/TUI docs.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `mission-workbench-tui`: Improve the empty-state UX so the default entry point guides incomplete profiles toward setup and verification instead of a dead-end prompt.
- `downstream-docs-sync`: Keep README and CLI/TUI docs aligned with the new workbench setup guidance.

## Impact

- Affected code: `internal/cli/cockpit/pages/missioncontrol.go`
- Affected tests: `internal/cli/workbench/model_test.go`, `internal/cli/cockpit/pages/missioncontrol_test.go`
- Affected docs: `README.md`, `docs/cli/core.md`, `docs/features/cockpit.md`
