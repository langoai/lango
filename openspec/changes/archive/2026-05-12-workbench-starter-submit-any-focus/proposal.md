## Why

The workbench already let the operator seed a starter prompt quickly, but once focus moved away from `Composer`, submitting that seeded starter still required a focus round-trip. That was unnecessary friction on the shortest path from launch to first useful result.

## What Changes

- Allow `Enter` to submit an already-armed starter prompt even when focus is not on the composer lane.
- Simplify seeded-state guidance so it no longer tells the operator to tab back to `Composer` first.
- Add regressions that prove seeded starters can be submitted from a non-composer focus state.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `mission-workbench-tui`: Seeded starter prompts can now be submitted from any empty-workbench focus lane, not only from `Composer`.

## Impact

- Affected code: `internal/cli/cockpit/pages/missioncontrol.go`
- Affected tests: `internal/cli/workbench/model_test.go`
