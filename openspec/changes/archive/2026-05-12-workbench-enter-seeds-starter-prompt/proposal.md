## Why

The workbench quick-start path already exposed starter prompts, but pressing `Enter` on the default empty first screen still did nothing. That left an unnecessary dead step in the launch-to-first-request flow.

## What Changes

- When the ready-profile workbench is empty, pressing `Enter` now seeds the first starter prompt.
- Preserve setup-first behavior for incomplete profiles.
- Update docs and regressions to reflect the shorter launch path.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `mission-workbench-tui`: Empty ready-profile workbench now treats `Enter` as a quick-start seed for the default starter prompt.
- `downstream-docs-sync`: Public docs now mention the `Enter` quick-start path.

## Impact

- Affected code: `internal/cli/cockpit/pages/missioncontrol.go`
- Affected tests: `internal/cli/workbench/model_test.go`
- Affected docs: `README.md`, `docs/cli/core.md`, `docs/features/cockpit.md`
