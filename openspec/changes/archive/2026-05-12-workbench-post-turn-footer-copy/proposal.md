## Why

The completed-turn workbench body and placeholder now describe a next-step loop, but the footer still says `Type to chat here`. That leaves one remaining copy seam out of sync in the same state.

## What Changes

- Change the completed-turn workbench footer from `Type to chat here` to `Type next prompt here`.
- Add regressions for the completed-turn footer wording.
- Sync public docs and specs for the refined footer copy.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `mission-workbench-tui`: the completed-turn footer now uses next-prompt wording instead of generic chat wording.
- `downstream-docs-sync`: public workbench docs describe the completed-turn footer as a next-prompt hint.

## Impact

- Affected code: `internal/cli/cockpit/pages/missioncontrol_workbench.go`
- Affected tests: `internal/cli/workbench/model_test.go`
- Affected docs: `README.md`, `docs/cli/core.md`, `docs/features/cockpit.md`
- Affected specs: `openspec/specs/mission-workbench-tui/spec.md`, `openspec/specs/downstream-docs-sync/spec.md`
