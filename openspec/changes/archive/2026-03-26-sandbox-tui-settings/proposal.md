## Why

The `sandbox.*` config (`internal/config/types_sandbox.go`) was added for OS-level tool execution isolation but has no TUI settings form. Users cannot configure sandbox settings through the interactive settings editor (`lango settings`). The existing P2P sandbox form uses `sandbox_*` field keys, so the new form must use a distinct `os_sandbox_*` namespace to prevent config mapping collisions.

## What Changes

- New TUI settings form `NewOSSandboxForm()` with 9 fields mapping to `cfg.Sandbox.*`
- Field keys use `os_sandbox_*` prefix to avoid collision with P2P sandbox's `sandbox_*` keys
- Menu entry in "Security" section, form dispatch, editor enabled check
- State update handlers for all 9 fields in `UpdateConfigFromForm()`
- Tests verifying field count, menu presence, and P2P config isolation

## Capabilities

### Modified Capabilities
- `cli-settings`: Added `os_sandbox` category to Security section with 9 configurable fields

## Impact

- **Code**: `internal/cli/settings/forms_sandbox.go` (new), `menu.go`, `setup_flow.go`, `editor.go`, `forms_impl_test.go`, `internal/cli/tuicore/state_update.go`
- **Config**: No schema changes — maps existing `SandboxConfig` to TUI
- **Risk**: Field key collision with P2P sandbox mitigated by `os_sandbox_*` namespace prefix
