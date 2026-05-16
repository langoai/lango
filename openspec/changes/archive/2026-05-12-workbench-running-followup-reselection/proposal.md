## Why

The running-state follow-up loop already taught that the operator could type the next prompt and interrupt the current turn with `Enter`, but the actual replacement path for staged follow-ups needed to stay just as direct. That interaction is worth treating as a first-class contract rather than an incidental side effect.

## What Changes

- Keep `1/2/3` active while a follow-up draft is staged during a running starter turn.
- Let those keys replace the staged follow-up directly instead of appending digits.
- Add a workbench regression proving the replacement path works during the running state.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `mission-workbench-tui`: Running-state follow-up drafts can now be replaced directly by starter hotkeys while the current turn is still in flight.

## Impact

- Affected code: `internal/cli/cockpit/pages/missioncontrol_workbench.go`
- Affected tests: `internal/cli/workbench/model_test.go`
