## Why

The workbench already supported `Enter` as a quick-start seed on the empty ready-profile screen, but the empty-state copy and footer still taught only the numeric hotkeys. That made a useful shortcut exist without being discoverable.

## What Changes

- Update ready-profile empty-state copy to mention `Enter` alongside `1/2/3`.
- Update the empty composer hint to explain that `Enter` loads the default starter prompt.
- Update the footer hint and docs to surface the same quick-start path consistently.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `mission-workbench-tui`: Ready-profile quick-start copy now advertises the `Enter` shortcut directly.
- `downstream-docs-sync`: Public docs now describe `Enter` as the default quick-start seed.

## Impact

- Affected code: `internal/cli/cockpit/pages/missioncontrol.go`
- Affected tests: `internal/cli/workbench/model_test.go`
- Affected docs: `README.md`, `docs/cli/core.md`, `docs/features/cockpit.md`
