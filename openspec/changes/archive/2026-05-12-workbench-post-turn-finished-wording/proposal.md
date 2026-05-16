## Why

The completed-turn empty workbench body currently says `Last turn complete`, which reads like a success-only statement even when the previous turn may have ended in failure. The next-step loop should be truthful about completion without implying success.

## What Changes

- Change the completed-turn empty body lead from `Last turn complete` to `Last turn finished`.
- Add regressions for the updated wording.
- Sync public docs and specs for the truth-aligned completed-turn wording.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `mission-workbench-tui`: the completed-turn empty body now uses neutral `finished` wording instead of success-leaning `complete`.
- `downstream-docs-sync`: public workbench docs describe the completed-turn state using neutral wording.

## Impact

- Affected code: `internal/cli/cockpit/pages/missioncontrol.go`
- Affected tests: `internal/cli/workbench/model_test.go`
- Affected docs: `README.md`, `docs/cli/core.md`, `docs/features/cockpit.md`
- Affected specs: `openspec/specs/mission-workbench-tui/spec.md`, `openspec/specs/downstream-docs-sync/spec.md`
