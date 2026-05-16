## Why

The standalone workbench can surface cockpit-style degraded warnings on its empty first screen even when those missing readers are optional for the current surface. That makes bare `lango` feel broken before the operator has taken any action.

## What Changes

- Suppress degraded-note rendering on the empty standalone workbench shell.
- Keep degraded-note rendering intact on the explicit cockpit surface.
- Sync public docs and specs for the calmer empty-state warning posture.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `mission-workbench-tui`: the empty standalone workbench no longer surfaces cockpit degraded warnings before any active mission/control content exists.
- `downstream-docs-sync`: public workbench docs describe the calmer empty-state warning posture.

## Impact

- Affected code: `internal/cli/cockpit/pages/missioncontrol.go`
- Affected tests: `internal/cli/cockpit/pages/missioncontrol_test.go`
- Affected docs: `README.md`, `docs/cli/core.md`, `docs/features/cockpit.md`
- Affected specs: `openspec/specs/mission-workbench-tui/spec.md`, `openspec/specs/downstream-docs-sync/spec.md`
