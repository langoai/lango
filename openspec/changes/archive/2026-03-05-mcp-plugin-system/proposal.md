## Why

Lango needs a way to extend agent capabilities by connecting to external MCP (Model Context Protocol) servers. Users should be able to configure external MCP servers (stdio, HTTP, SSE) and have their tools automatically available to the agent — matching the extensibility model used by Claude Code and other MCP-compatible clients.

## What Changes

- New `MCPConfig` / `MCPServerConfig` config types with multi-scope loading (profile < user < project)
- New `internal/mcp/` package: connection lifecycle, server manager, tool adapter, env expansion, config file loading
- App wiring: MCP tools injected into the agent tool pipeline with full middleware chain (approval, hooks, learning)
- Lifecycle management: MCP connections gracefully shut down on app stop
- CLI commands: `lango mcp add|remove|list|get|test|enable|disable`
- Secret scanning: MCP server auth headers registered with the secret scanner
- Exec guard: `lango mcp` blocked from agent shell execution

## Capabilities

### New Capabilities
- `mcp-integration`: MCP server connection management, tool discovery, tool adaptation, health checks, auto-reconnect, multi-scope config

### Modified Capabilities
- `tool-exec`: Added `lango mcp` to the `blockLangoExec` guard list

## Impact

- **Config**: `internal/config/types.go` (new MCP field), `internal/config/types_mcp.go` (new file), `internal/config/loader.go` (defaults, validation, env substitution)
- **Core**: `internal/mcp/` (new package: errors, env, connection, manager, adapter, config_loader)
- **App**: `internal/app/types.go` (MCPManager field), `internal/app/app.go` (init sequence, lifecycle, secrets), `internal/app/wiring_mcp.go`, `internal/app/tools.go` (exec guard)
- **CLI**: `internal/cli/mcp/` (new package: 7 subcommands), `cmd/lango/main.go` (registration)
- **Dependencies**: `github.com/modelcontextprotocol/go-sdk` v1.4.0 added
