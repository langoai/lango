## Why

The TUI Settings (`lango settings`) MCP menu currently only allows editing global MCP settings (enabled, timeout, etc.). Individual MCP server management (add/edit/delete) is only possible via CLI (`lango mcp add/remove`). Users need a consistent CRUD interface for MCP servers within the TUI, following the existing Providers and Auth Providers list+form pattern.

## What Changes

- Add `MCPServersListModel` for browsing, selecting, and deleting MCP servers in TUI
- Add `NewMCPServerForm()` for adding/editing individual server configurations with transport-conditional fields
- Add `UpdateMCPServerFromForm()` to persist form data back to `cfg.MCP.Servers`
- Add `StepMCPServersList` editor step with full navigation wiring
- Split the Infrastructure menu: "MCP Settings" (global) + "MCP Server List" (CRUD)

## Capabilities

### New Capabilities
- `tui-mcp-server-crud`: TUI list+form CRUD for individual MCP server configurations

### Modified Capabilities
- `tui-mcp-settings`: Menu label changed from "MCP Servers" to "MCP Settings" to disambiguate from the new server list entry

## Impact

- `internal/cli/settings/mcp_servers_list.go` — new file
- `internal/cli/settings/forms_mcp.go` — new `NewMCPServerForm`, `formatKeyValuePairs`
- `internal/cli/tuicore/state_update.go` — new `UpdateMCPServerFromForm`, `parseKeyValuePairs`
- `internal/cli/settings/editor.go` — new step, wiring, form detection
- `internal/cli/settings/menu.go` — split MCP menu entry
