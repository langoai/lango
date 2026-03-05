## Why

MCP Plugin System (Phase 1-4) implementation is complete but the TUI Settings editor has no form for MCP global configuration. All other major features (Cron, P2P, KMS, etc.) have corresponding TUI settings forms. Per CLAUDE.md rules, core code changes must include UI/UX updates in the same response.

## What Changes

- Add `NewMCPForm()` with 6 fields: Enabled, Default Timeout, Max Output Tokens, Health Check Interval, Auto Reconnect, Max Reconnect Attempts
- Add "MCP Servers" category to the Infrastructure section in the settings menu
- Add `case "mcp"` handler in the editor's menu selection dispatcher
- Add MCP field update cases in `UpdateConfigFromForm()` state binding

## Capabilities

### New Capabilities
- `tui-mcp-settings`: TUI settings form for MCP global configuration (enabled, timeouts, reconnection)

### Modified Capabilities
- `mcp-integration`: Adding TUI settings surface for existing MCP config fields

## Impact

- `internal/cli/settings/forms_mcp.go` — new file
- `internal/cli/settings/menu.go` — Infrastructure section gains MCP entry
- `internal/cli/settings/editor.go` — new case in `handleMenuSelection()`
- `internal/cli/tuicore/state_update.go` — 6 new cases in `UpdateConfigFromForm()`
