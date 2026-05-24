## Why

The workbench empty state already distinguishes incomplete and ready profiles, but the header summary could still imply readiness by showing only the provider ID. That creates a false signal at the top of the first screen even when setup is not actually usable yet.

## What Changes

- Mark incomplete workbench profiles as `Setup required` in the header model/provider summary.
- Preserve the normal `provider / model` summary once the active profile is actually usable.
- Update workbench docs and tests so the readiness signal stays aligned across header, empty state, and composer guidance.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `mission-workbench-tui`: Header summary now reflects configuration readiness instead of echoing a misleading provider-only default.
- `downstream-docs-sync`: Workbench docs mention the readiness-aware header summary.

## Impact

- Affected code: `internal/cli/cockpit/missioncontrol_projector.go`
- Affected tests: `internal/cli/cockpit/missioncontrol_projector_test.go`, `internal/cli/workbench/model_test.go`
- Affected docs: `README.md`, `docs/cli/core.md`, `docs/features/cockpit.md`
