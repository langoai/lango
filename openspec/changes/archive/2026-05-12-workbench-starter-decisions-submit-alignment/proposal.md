## Why

The workbench spec already said that an armed starter prompt should submit from any empty-workbench focus lane, but the implementation still missed the `Decisions` lane. That left the code behind the documented contract.

## What Changes

- Let an armed starter prompt submit from the `Decisions` focus lane as well as `Missions` and `Composer`.
- Add focused regressions so that lane-independent submit behavior stays aligned with the spec.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `mission-workbench-tui`: The implementation now matches the documented any-focus seeded-starter submit contract.

## Impact

- Affected code: `internal/cli/cockpit/pages/missioncontrol.go`
- Affected tests: `internal/cli/workbench/model_test.go`, `internal/cli/cockpit/pages/missioncontrol_test.go`
