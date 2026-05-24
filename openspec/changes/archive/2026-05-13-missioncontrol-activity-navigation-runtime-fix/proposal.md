## Why

Mission Control already advertises `↑/↓` activity navigation when the composer lane is focused and multiple activity rows exist, but the current key routing forwards those keys into the composer instead of moving the activity cursor. That leaves the help bar and docs promising a navigation path that does not actually work.

## What Changes

- Route `↑/k` and `↓/j` to the activity cursor when the composer/activity lane is focused and multiple activity rows exist.
- Add regression coverage proving the activity cursor moves in response to those keys.
- Sync the cockpit-pages spec and cockpit feature docs with the runtime activity-navigation contract.

## Capabilities

### New Capabilities

### Modified Capabilities

- `cockpit-pages`: Mission Control activity-lane navigation keys need an explicit runtime behavior contract.

## Impact

- Affected code: `internal/cli/cockpit/pages/missioncontrol.go`, `internal/cli/cockpit/pages/missioncontrol_test.go`
- Affected docs/specs: `docs/features/cockpit.md`, `openspec/specs/cockpit-pages/spec.md`
