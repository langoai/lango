## Why

The workbench UI already tells the operator they can edit an armed starter prompt before sending, but editing keys like `Backspace` did not work unless focus was still on `Composer`. That left a subtle mismatch between the copy and the actual behavior.

## What Changes

- Route composer editing keys to the armed starter prompt even when focus has moved away from `Composer`.
- Move focus back to `Composer` automatically when that editing path is used.
- Add a workbench regression to lock in the behavior.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `mission-workbench-tui`: Armed starter prompts can now be edited immediately with composer editing keys even after focus leaves `Composer`.

## Impact

- Affected code: `internal/cli/cockpit/pages/missioncontrol.go`
- Affected tests: `internal/cli/workbench/model_test.go`
