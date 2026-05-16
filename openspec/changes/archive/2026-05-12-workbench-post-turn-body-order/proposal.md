## Why

The completed-turn workbench body now contains the right ingredients, but it still shows the generic next-prompt hint before the primary next-step starter guidance. That makes the recommended action visually secondary.

## What Changes

- Reorder the completed-turn empty workbench body so the recommended next-step starter appears before the generic next-prompt hint.
- Add regressions for the completed-turn body ordering.
- Sync the mission-workbench spec for the ordering contract.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `mission-workbench-tui`: the completed-turn empty body now presents the primary next-step starter before the secondary typing hint.

## Impact

- Affected code: `internal/cli/cockpit/pages/missioncontrol.go`
- Affected tests: `internal/cli/workbench/model_test.go`
- Affected specs: `openspec/specs/mission-workbench-tui/spec.md`
