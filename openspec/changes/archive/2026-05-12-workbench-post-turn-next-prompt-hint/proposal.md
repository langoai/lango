## Why

The completed-turn empty workbench now uses a next-step default starter, but its chat hint still says `Type to chat here`. That wording is generic and misses the fact that the surface is now steering the operator into the next prompt loop.

## What Changes

- Make the completed-turn empty workbench hint say `Type the next prompt here`.
- Add regressions for the completed-turn hint path.
- Sync public docs and specs for the next-prompt wording.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `mission-workbench-tui`: the completed-turn empty workbench hint now explicitly invites the operator to type the next prompt.
- `downstream-docs-sync`: public workbench docs describe the completed-turn hint as a next-prompt loop.

## Impact

- Affected code: `internal/cli/cockpit/pages/missioncontrol.go`
- Affected tests: `internal/cli/workbench/model_test.go`
- Affected docs: `README.md`, `docs/cli/core.md`, `docs/features/cockpit.md`
- Affected specs: `openspec/specs/mission-workbench-tui/spec.md`, `openspec/specs/downstream-docs-sync/spec.md`
