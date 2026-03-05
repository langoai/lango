## 1. MCP Servers List Model

- [x] 1.1 Create `MCPServerItem` struct and `MCPServersListModel` in `internal/cli/settings/mcp_servers_list.go`
- [x] 1.2 Implement `NewMCPServersListModel(cfg)` — build items from `cfg.MCP.Servers`, sort by name
- [x] 1.3 Implement `Update()` — up/down/enter/d/esc key handling
- [x] 1.4 Implement `View()` — render items as `name (transport) [enabled/disabled]` with "+ Add New MCP Server"

## 2. MCP Server Form

- [x] 2.1 Create `NewMCPServerForm(name, srv)` in `internal/cli/settings/forms_mcp.go`
- [x] 2.2 Add transport-conditional field visibility via `VisibleWhen` closures
- [x] 2.3 Add `formatKeyValuePairs()` helper for map→CSV serialization

## 3. State Update

- [x] 3.1 Add `UpdateMCPServerFromForm(name, form)` in `internal/cli/tuicore/state_update.go`
- [x] 3.2 Add `parseKeyValuePairs()` helper for CSV→map deserialization

## 4. Editor Integration

- [x] 4.1 Add `StepMCPServersList` constant and `mcpServersList`/`activeMCPServerName` fields to `Editor`
- [x] 4.2 Add `mcp_servers` case to `handleMenuSelection()` — initialize list and set step
- [x] 4.3 Add `StepMCPServersList` case to `Update()` — handle delete/select/exit/new
- [x] 4.4 Update `StepForm` Esc handler — detect MCP server forms, call `UpdateMCPServerFromForm`, return to list
- [x] 4.5 Add `isMCPServerForm()` helper
- [x] 4.6 Update `View()` — add breadcrumb and content rendering for `StepMCPServersList`

## 5. Menu Update

- [x] 5.1 Split MCP menu entry: "MCP Settings" (global) + "MCP Server List" (CRUD)

## 6. Verification

- [x] 6.1 Run `go build ./...` — verify clean compilation
- [x] 6.2 Run `go test ./internal/cli/settings/... ./internal/cli/tuicore/...` — all tests pass
