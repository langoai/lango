## Why

Once a starter prompt is seeded, the next action depends on focus. If focus has left the composer lane, telling the operator only to press `Enter` is incomplete guidance.

## What Changes

- Make seeded-starter body copy mention `Tab` back to `Composer` before `Enter` when focus is not on the composer lane.
- Make the footer hint reflect the same focus-aware guidance.
- Add a workbench regression to keep that copy aligned with actual key behavior.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `mission-workbench-tui`: Seeded-starter guidance now reflects whether the composer lane still has focus.

## Impact

- Affected code: `internal/cli/cockpit/pages/missioncontrol.go`
- Affected tests: `internal/cli/workbench/model_test.go`
