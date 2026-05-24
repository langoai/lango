## Why

The workbench empty state already distinguishes incomplete profiles from ready ones, but the composer placeholder still used generic copy. That split-brain UX weakens the first action cue right at the point where the operator starts typing.

## What Changes

- Make the workbench composer placeholder follow the same incomplete-vs-ready guidance as the empty-state body.
- Keep cockpit behavior unchanged; only the workbench-flavored Mission Control surface gets the state-aware composer hint.
- Update docs and specs so the workbench contract includes state-aware composer guidance.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `mission-workbench-tui`: Extend state-aware workbench guidance into the composer placeholder.
- `downstream-docs-sync`: Keep workbench docs aligned with the composer-guidance split.

## Impact

- Affected code: `internal/cli/cockpit/pages/missioncontrol.go`
- Affected tests: `internal/cli/workbench/model_test.go`
- Affected docs/specs: `README.md`, `docs/cli/core.md`, `docs/features/cockpit.md`
