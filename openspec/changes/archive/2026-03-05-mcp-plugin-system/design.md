## Architecture

### Layer Placement

MCP integration is a **Network-layer** component. It sits between the config system and the tool pipeline:

```
Config (MCPConfig) → MCP Package (connection/manager/adapter) → App Wiring → Tool Pipeline
```

- **Core boundary**: `internal/mcp/` depends only on `internal/config`, `internal/agent`, `internal/logging`
- **App boundary**: `internal/app/wiring_mcp.go` bridges MCP into the init sequence
- **CLI boundary**: `internal/cli/mcp/` uses only config + `internal/mcp` (no app dependency)

### Connection Model

Each MCP server is managed by a `ServerConnection` that wraps:
1. Transport creation (stdio → `CommandTransport`, http → `StreamableClientTransport`, sse → `SSEClientTransport`)
2. Client/session lifecycle via `mcp.NewClient().Connect()`
3. Capability discovery (tools, resources, prompts) via SDK iterators
4. Health check goroutine with `session.Ping()` + exponential backoff reconnection

`ServerManager` orchestrates multiple connections with concurrent connect/disconnect.

### Tool Adaptation

MCP tools are converted to `agent.Tool` using the naming convention `mcp__{serverName}__{toolName}`:
- `InputSchema` (any) → `agent.Tool.Parameters` (map extraction)
- `SafetyLevel` from server config (default: Dangerous)
- Handler proxies to `session.CallTool()` with per-server timeout
- Output truncation at configurable max tokens (default: 25000)

### Multi-Scope Config

Three config sources merged in priority order:
1. Profile config (config DB, lowest priority)
2. User-level: `~/.lango/mcp.json`
3. Project-level: `.lango-mcp.json` (highest priority, committable to VCS)

All support `${VAR}` and `${VAR:-default}` environment variable expansion.

## Key Decisions

| Decision | Rationale |
|----------|-----------|
| Client-only (no MCP server hosting) | Lango consumes MCP servers; A2A serves the outbound role |
| Tool naming `mcp__{server}__{tool}` | Matches Claude Code convention; prevents name collisions |
| Default safety = Dangerous | Fail-safe for untrusted external tools |
| MCP tools go through full middleware chain | Same approval, hooks, learning as built-in tools |
| Health check per connection goroutine | Simple, reliable; stop channel for clean shutdown |
| MCP init after dispatcher tools, before hooks | Tools receive full middleware wrapping |

## Dependencies

- `github.com/modelcontextprotocol/go-sdk` v1.4.0 (Go MCP SDK, MIT license)
- No new internal package dependencies beyond existing patterns
