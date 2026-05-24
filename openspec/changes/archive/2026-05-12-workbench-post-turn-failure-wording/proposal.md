## Why

The completed-turn workbench now uses neutral `finished` wording, but failure turns still read almost the same as successful turns in the empty body. The latest summary is visible, yet the primary lead line does not explicitly call out that the last turn needs attention.

## What Changes

- Make the completed-turn empty body switch to an attention-oriented lead when the latest assistant summary represents a failed turn.
- Add regressions for the failure wording path.
- Sync public docs and specs for the failure-aware completed-turn wording.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `mission-workbench-tui`: completed-turn empty-body wording now distinguishes failed turns from successful turns.
- `downstream-docs-sync`: public workbench docs mention the failure-aware completed-turn wording.

## Impact

- Affected code: `internal/cli/cockpit/pages/missioncontrol.go`
- Affected tests: `internal/cli/workbench/model_test.go`
- Affected docs: `README.md`, `docs/cli/core.md`, `docs/features/cockpit.md`
- Affected specs: `openspec/specs/mission-workbench-tui/spec.md`, `openspec/specs/downstream-docs-sync/spec.md`
