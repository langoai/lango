## Why

The workbench already documented an interrupt-and-run loop for staged follow-ups, but the implementation still missed submission from the `Decisions` and `Missions` lanes while a follow-up draft was staged. That left the code behind the interaction contract.

## What Changes

- Route `Enter` from non-composer empty-workbench lanes to the staged follow-up interrupt path when a follow-up draft exists.
- Add workbench regressions proving that a staged follow-up can be submitted from non-composer focus.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `mission-workbench-tui`: Running-state staged follow-ups now submit from non-composer empty-workbench lanes too, matching the interaction contract already taught by the UI.

## Impact

- Affected code: `internal/cli/cockpit/pages/missioncontrol.go`
- Affected tests: `internal/cli/workbench/model_test.go`
