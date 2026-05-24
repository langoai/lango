## Why

The workbench now changes its post-turn default starter, but the empty-state copy still talks like the operator is on the very first quick-start step. That makes the behavior correct but the copy slightly misleading.

## What Changes

- Change completed-turn empty-state workbench copy from `Quick start` / `Enter default starter` framing to `Next step` / `Enter next-step starter` framing.
- Add regressions for the post-turn copy path.
- Sync public docs and specs so the wording matches the actual post-turn loop.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `mission-workbench-tui`: post-turn empty-state copy now explicitly frames the default starter as the next step instead of the initial quick start.
- `downstream-docs-sync`: public workbench docs describe the post-turn copy as a next-step default.

## Impact

- Affected code: `internal/cli/cockpit/pages/missioncontrol_workbench.go`
- Affected tests: `internal/cli/workbench/model_test.go`
- Affected docs: `README.md`, `docs/cli/core.md`, `docs/features/cockpit.md`
- Affected specs: `openspec/specs/mission-workbench-tui/spec.md`, `openspec/specs/downstream-docs-sync/spec.md`
