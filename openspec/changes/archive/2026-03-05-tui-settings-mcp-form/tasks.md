## 1. Form Implementation

- [x] 1.1 Create `internal/cli/settings/forms_mcp.go` with `NewMCPForm(cfg *config.Config)` returning 6 fields (enabled, default_timeout, max_output_tokens, health_check_interval, auto_reconnect, max_reconnect_attempts)
- [x] 1.2 Add duration validation (`time.ParseDuration`) on timeout and interval fields
- [x] 1.3 Add positive integer validation on max_output_tokens and max_reconnect_attempts fields

## 2. Menu & Editor Wiring

- [x] 2.1 Add `{"mcp", "MCP Servers", "External MCP server integration"}` to Infrastructure section in `menu.go`
- [x] 2.2 Add `case "mcp"` handler in `editor.go` `handleMenuSelection()` to open `NewMCPForm`

## 3. Config State Binding

- [x] 3.1 Add 6 MCP cases in `tuicore/state_update.go` `UpdateConfigFromForm()`: mcp_enabled, mcp_default_timeout, mcp_max_output_tokens, mcp_health_check_interval, mcp_auto_reconnect, mcp_max_reconnect_attempts

## 4. Verification

- [x] 4.1 Run `go build ./...` — no compilation errors
- [x] 4.2 Run `go test ./internal/cli/settings/...` — all tests pass
- [x] 4.3 Run `go test ./internal/cli/tuicore/...` — all tests pass
