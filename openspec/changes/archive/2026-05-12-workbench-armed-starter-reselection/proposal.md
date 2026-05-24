## Why

Once a starter prompt is armed, the workbench should not force the operator to clear it manually just to choose another starter. The seeded state is part of the quick-start loop, so replacement should stay just as direct as the initial selection.

## What Changes

- Keep `1/2/3` active even after a starter prompt is already armed.
- Let those keys replace the armed starter prompt directly instead of appending text.
- Add a regression proving the replacement path works.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `mission-workbench-tui`: Armed starter prompts can now be replaced directly with the numeric starter shortcuts.

## Impact

- Affected code: `internal/cli/cockpit/pages/missioncontrol.go`
- Affected tests: `internal/cli/workbench/model_test.go`
