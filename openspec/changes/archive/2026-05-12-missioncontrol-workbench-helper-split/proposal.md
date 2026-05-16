## Why

The Mission Control page file kept accumulating workbench-specific setup and quick-start helpers. Even with correct behavior, that concentration makes the file harder to reason about and raises the cost of future changes.

## What Changes

- Split workbench-specific setup, starter-prompt, and seeded-state helpers into a dedicated page-adjacent source file.
- Preserve all existing workbench behavior and tests.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `mission-workbench-tui`: Workbench helper behavior is now isolated in a dedicated file, reducing responsibility sprawl inside the core Mission Control page source.

## Impact

- Affected code: `internal/cli/cockpit/pages/missioncontrol_workbench.go`, `internal/cli/cockpit/pages/missioncontrol.go`
- Affected tests: `internal/cli/workbench`, `internal/cli/cockpit/pages`
