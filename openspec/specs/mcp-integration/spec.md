# MCP Integration

## Purpose

Enable Lango to connect to external MCP (Model Context Protocol) servers and expose their tools to the agent.

## Requirements

### Configuration

- MUST support `mcp.enabled` boolean flag (default: false)
- MUST support named server configs under `mcp.servers.<name>`
- Each server MUST specify transport type: `stdio`, `http`, or `sse`
- Stdio servers MUST have `command`; http/sse servers MUST have `url`
- MUST support `${VAR}` and `${VAR:-default}` env var expansion in `env` and `headers`
- MUST support per-server `enabled` toggle (default: true)
- MUST support per-server `timeout` override
- MUST support per-server `safetyLevel`: safe, moderate, dangerous (default: dangerous)
- MUST support global `defaultTimeout` (30s), `maxOutputTokens` (25000), `healthCheckInterval` (30s)
- MUST merge configs from three scopes: profile < user (`~/.lango/mcp.json`) < project (`.lango-mcp.json`)

### Connection Lifecycle

- MUST connect to all enabled servers during app initialization
- MUST handle connection failures gracefully (log warning, continue with available servers)
- MUST support health checks via periodic `Ping()` with configurable interval
- MUST auto-reconnect on failure with exponential backoff (capped at 30s)
- MUST disconnect all servers on app shutdown via lifecycle registry (PriorityNetwork)

### Tool Adaptation

- MUST name adapted tools as `mcp__{serverName}__{toolName}`
- MUST convert MCP `InputSchema` to `agent.Tool.Parameters`
- MUST apply server-configured safety level to all adapted tools
- MUST proxy tool calls through `session.CallTool()` with timeout
- MUST truncate output exceeding `maxOutputTokens` (approximate: 4 chars/token)
- MUST pass MCP tools through the full middleware chain (hooks, approval, learning)

### Management Tools

- MUST provide `mcp_status` tool showing server connection states
- MUST provide `mcp_tools` tool listing available MCP tools (with optional server filter)
- MUST register MCP tools in tool catalog under "mcp" category

### CLI

- MUST provide `lango mcp list` to show configured servers
- MUST provide `lango mcp add <name>` with transport, command/url, env, headers, scope flags
- MUST provide `lango mcp remove <name>` to delete a server config
- MUST provide `lango mcp get <name>` to show server details and discovered tools
- MUST provide `lango mcp test <name>` to verify connectivity (handshake + ping + tool count)
- MUST provide `lango mcp enable/disable <name>` to toggle servers
- MUST support `--scope user|project` for add/remove/enable/disable operations

### Security

- MUST register MCP server auth headers with the secret scanner
- MUST block `lango mcp` from agent shell execution via `blockLangoExec` guard

## Scenarios

### Happy Path: Stdio Server
1. User configures `mcp.enabled: true` with a stdio server
2. App starts, connects to server, discovers tools
3. Agent can invoke MCP tools with `mcp__{server}__{tool}` naming
4. Health checks maintain connection; auto-reconnect on failure
5. App shutdown disconnects cleanly

### Happy Path: HTTP Server
1. User adds HTTP server via `lango mcp add api --type http --url https://...`
2. Server config saved to `~/.lango/mcp.json`
3. `lango mcp test api` verifies connectivity
4. On next `lango serve`, HTTP MCP tools are available

### Error: Connection Failure
1. Configured server is unreachable
2. Connection attempt fails with warning log
3. Other servers connect normally
4. Auto-reconnect attempts in background (if enabled)

### Multi-Scope Config
1. Team commits `.lango-mcp.json` with shared servers
2. Individual user adds personal server to `~/.lango/mcp.json`
3. Both sets of servers are available, project scope overrides on name conflicts
