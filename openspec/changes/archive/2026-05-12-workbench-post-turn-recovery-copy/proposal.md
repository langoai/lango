## Why

The completed-turn workbench now calls out failed turns in the body lead, but the starter/body/footer wording that follows still uses the same generic next-step language as successful turns. That misses a chance to orient the operator toward recovery.

## What Changes

- Switch completed-turn starter/body/footer wording to `Recovery step` / `Enter recovery starter` when the latest turn failed.
- Add regressions for the failure-aware recovery wording.
- Sync public docs and specs for the recovery-specific copy.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `mission-workbench-tui`: failed completed-turn states now use recovery-specific starter/body/footer wording.
- `downstream-docs-sync`: public workbench docs mention recovery wording after failed turns.

## Impact

- Affected code: `internal/cli/cockpit/pages/missioncontrol.go`, `internal/cli/cockpit/pages/missioncontrol_workbench.go`
- Affected tests: `internal/cli/workbench/model_test.go`
- Affected docs: `README.md`, `docs/cli/core.md`, `docs/features/cockpit.md`
- Affected specs: `openspec/specs/mission-workbench-tui/spec.md`, `openspec/specs/downstream-docs-sync/spec.md`
