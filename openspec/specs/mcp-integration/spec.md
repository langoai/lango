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
- MUST treat missing user/project scoped config files as absent optional configuration
- MUST return an actionable error for present user/project scoped config files that cannot be read, parsed, or validated

#### Scenario: Configuration merges all supported scopes
- **WHEN** MCP configuration is loaded from profile, user, and project scopes
- **THEN** the system MUST merge those scopes in the documented order
- **AND** preserve per-server transport, timeout, safety, and auth settings

#### Scenario: Missing scoped MCP config files are optional
- **WHEN** the user or project MCP config file does not exist
- **THEN** the system SHALL continue loading configuration from the remaining scopes
- **AND** it SHALL NOT return an error for the missing optional file

#### Scenario: Invalid scoped MCP config fails visibly
- **WHEN** a user or project MCP config file exists but cannot be read, parsed, or validated
- **THEN** the system SHALL return an error
- **AND** the error SHALL identify the scope and file path
- **AND** the system SHALL NOT silently fall back to lower-priority MCP configuration

#### Scenario: MCP CLI write commands preserve invalid existing files
- **WHEN** an MCP CLI write command targets a scope whose config file exists but is invalid
- **THEN** the command SHALL return an actionable config load error
- **AND** it SHALL NOT overwrite that existing file with replacement content

#### Scenario: MCP configuration merges documented scopes
- **WHEN** MCP configuration is loaded from profile, user, and project scopes
- **THEN** the system MUST merge those scopes in the documented order
- **AND** preserve per-server transport, timeout, safety, and auth settings

### Connection Lifecycle

- MUST connect to all enabled servers during app initialization
- MUST handle connection failures gracefully (log warning, continue with available servers)
- MUST support health checks via periodic `Ping()` with configurable interval
- MUST auto-reconnect on failure with exponential backoff (capped at 30s)
- MUST disconnect all servers on app shutdown via lifecycle registry (PriorityNetwork)

#### Scenario: Connection failure does not block other servers
- **WHEN** one enabled MCP server fails to connect during initialization
- **THEN** the system MUST log a warning
- **AND** continue with the remaining available servers

#### Scenario: Connection failure does not block remaining servers
- **WHEN** one enabled MCP server fails to connect during initialization
- **THEN** the system MUST log a warning
- **AND** continue with the remaining available servers

### Tool Adaptation

- MUST name adapted tools as `mcp__{serverName}__{toolName}`
- MUST convert MCP `InputSchema` to `agent.Tool.Parameters`
- MUST apply server-configured safety level to all adapted tools
- MUST proxy tool calls through `session.CallTool()` with timeout
- MUST truncate output exceeding `maxOutputTokens` (approximate: 4 chars/token)
- MUST pass MCP tools through the full middleware chain (hooks, approval, learning)

#### Scenario: MCP tools are adapted into agent tools
- **WHEN** tools are discovered from a connected MCP server
- **THEN** they MUST be exposed as `mcp__{serverName}__{toolName}`
- **AND** preserve schema, safety level, timeout, truncation, and middleware behavior

#### Scenario: MCP tools are adapted into agent tools
- **WHEN** tools are discovered from a connected MCP server
- **THEN** they MUST be exposed as `mcp__{serverName}__{toolName}`
- **AND** preserve schema, safety level, timeout, truncation, and middleware behavior

### Management Tools

- MUST provide `mcp_status` tool showing server connection states
- MUST provide `mcp_tools` tool listing available MCP tools (with optional server filter)
- MUST register MCP tools in tool catalog under "mcp" category

#### Scenario: Management tools expose server and tool inventory
- **WHEN** an operator queries MCP management tools
- **THEN** `mcp_status` MUST report server connection states
- **AND** `mcp_tools` MUST list available MCP tools with optional server filtering

#### Scenario: Management tools expose server and tool inventory
- **WHEN** an operator queries MCP management tools
- **THEN** `mcp_status` MUST report server connection states
- **AND** `mcp_tools` MUST list available MCP tools with optional server filtering

### CLI

- MUST provide `lango mcp list` to show configured servers
- MUST provide `lango mcp add <name>` with transport, command/url, env, headers, scope flags
- MUST provide `lango mcp remove <name>` to delete a server config
- MUST provide `lango mcp get <name>` to show server details and discovered tools
- MUST provide `lango mcp test <name>` to verify connectivity (handshake + ping + tool count)
- MUST provide `lango mcp enable/disable <name>` to toggle servers
- MUST support `--scope user|project` for add/remove/enable/disable operations

#### Scenario: CLI manages MCP server lifecycle
- **WHEN** an operator uses `lango mcp` subcommands
- **THEN** the CLI MUST support listing, adding, removing, inspecting, testing, enabling, and disabling servers
- **AND** apply scope-aware mutations where supported

#### Scenario: MCP get output uses the command writer
- **WHEN** an operator runs `lango mcp get <name>`
- **THEN** the command SHALL write the full output through the Cobra command output writer
- **AND** wrappers or tests that replace `cmd.OutOrStdout()` SHALL capture the command output

#### Scenario: MCP list output uses the command writer
- **WHEN** an operator runs `lango mcp list`
- **THEN** the command SHALL write the full output through the Cobra command output writer
- **AND** wrappers or tests that replace `cmd.OutOrStdout()` SHALL capture the command output

#### Scenario: MCP enable/disable output uses the command writer
- **WHEN** an operator runs `lango mcp enable <name>` or `lango mcp disable <name>`
- **THEN** the command SHALL write the full confirmation output through the Cobra command output writer
- **AND** wrappers or tests that replace `cmd.OutOrStdout()` SHALL capture the command output

#### Scenario: MCP add/remove output uses the command writer
- **WHEN** an operator runs `lango mcp add <name>` or `lango mcp remove <name>`
- **THEN** the command SHALL write the full confirmation output through the Cobra command output writer
- **AND** wrappers or tests that replace `cmd.OutOrStdout()` SHALL capture the command output

#### Scenario: MCP test output uses the command writer
- **WHEN** an operator runs `lango mcp test <name>`
- **THEN** the command SHALL write the full diagnostic output through the Cobra command output writer
- **AND** wrappers or tests that replace `cmd.OutOrStdout()` SHALL capture the command output

#### Scenario: CLI manages MCP server lifecycle
- **WHEN** an operator uses `lango mcp` subcommands
- **THEN** the CLI MUST support listing, adding, removing, inspecting, testing, enabling, and disabling servers
- **AND** apply scope-aware mutations where supported

### TUI Settings

- MCP integration SHALL be configurable through both CLI commands and the TUI settings editor
- Global settings (enabled, timeouts, reconnection) SHALL be available in the TUI settings form under Infrastructure > MCP Servers
- Individual server management (add/remove/enable/disable) SHALL remain CLI-only via `lango mcp` subcommands

#### Scenario: TUI exposes global MCP settings only
- **WHEN** an operator uses the settings editor for MCP configuration
- **THEN** the TUI SHALL expose global MCP settings
- **AND** per-server lifecycle actions SHALL remain CLI-only

#### Scenario: TUI exposes global MCP settings only
- **WHEN** an operator uses the settings editor for MCP configuration
- **THEN** the TUI SHALL expose global MCP settings
- **AND** per-server lifecycle actions SHALL remain CLI-only

### Security

- MUST register MCP server auth headers with the secret scanner
- MUST block `lango mcp` from agent shell execution via `blockLangoExec` guard

#### Scenario: MCP secrets remain behind scanner and exec guard
- **WHEN** MCP auth headers are configured and shell execution is evaluated
- **THEN** the secret scanner MUST register the headers
- **AND** `blockLangoExec` MUST block agent shell execution of `lango mcp`

#### Scenario: MCP secrets remain behind scanner and exec guard
- **WHEN** MCP auth headers are configured and shell execution is evaluated
- **THEN** the secret scanner MUST register the headers
- **AND** `blockLangoExec` MUST block agent shell execution of `lango mcp`

### Requirement: MCP stdio server OS sandbox
The MCP `ServerConnection` SHALL support optional OS-level sandbox for stdio server processes via `SetOSIsolator(iso, dataRoot)`, applied at transport creation time with `MCPServerPolicy(dataRoot)` (network=allow, filesystem restricted, lango control-plane denied).

`SetOSIsolator` SHALL accept a `dataRoot string` second argument so the policy applied at transport creation time denies the lango control-plane (`~/.lango`) to the spawned MCP server child process. Empty `dataRoot` skips the control-plane mask (used by unit tests).

`ServerConnection` SHALL also accept an optional `bus *eventbus.Bus` via `SetEventBus(bus)`. When set, the connection SHALL publish a `SandboxDecisionEvent` with `Source="mcp"`, `Command=sc.name`, and empty `SessionKey` (MCP server lifecycle is process-level) for every decision branch in `createTransport`: `applied`, `skipped`, `rejected` (including the `failClosed && isolator==nil` rejection path), so the audit trail records both successful and failed sandbox decisions for MCP transports.

`mcp.ServerManager` SHALL provide `SetOSIsolator(iso, dataRoot)` and `SetEventBus(bus)` methods that store both values on the manager AND propagate them to every existing connection. `ConnectAll` SHALL pass the manager's `dataRoot` and `bus` to each newly-created connection.

#### Scenario: Stdio server sandboxed with control-plane mask
- **WHEN** an MCP stdio server is started with isolator and dataRoot configured
- **THEN** the server process SHALL run with filesystem restrictions (read-global, write-/tmp only, lango control-plane denied) while retaining network access

#### Scenario: Sandbox error is non-fatal
- **WHEN** the isolator returns an error during transport creation and `failClosed=false`
- **THEN** the server SHALL start without sandbox, log a warning, and publish `SandboxDecisionEvent{Source:"mcp", Decision:"skipped"}`

#### Scenario: MCP decision events have empty SessionKey
- **WHEN** `createTransport()` publishes a `SandboxDecisionEvent`
- **THEN** the `SessionKey` field SHALL be empty because MCP server lifecycle is process-level, not session-bound
- **AND** the audit recorder SHALL accept the row with no session key set

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

### Documentation

#### Requirement: MCP documentation coverage
The MCP Plugin System SHALL have complete documentation coverage across README.md and docs/cli/ matching all other documented features.

#### Scenario: README Features list includes MCP
- **WHEN** a user reads the README.md Features section
- **THEN** MCP Integration is listed with description of stdio/HTTP/SSE transport, auto-discovery, health checks, and multi-scope config

#### Scenario: README CLI Commands section includes MCP
- **WHEN** a user reads the README.md CLI Commands section
- **THEN** all 7 `lango mcp` subcommands (list, add, remove, get, test, enable, disable) are listed with descriptions

#### Scenario: README Architecture diagram includes MCP
- **WHEN** a user reads the README.md Architecture section
- **THEN** `mcp/` appears in both the cli/ tree and the internal/ tree

#### Scenario: docs/cli/index.md Quick Reference includes MCP
- **WHEN** a user reads the CLI Quick Reference table in docs/cli/index.md
- **THEN** an "MCP Servers" section lists all 7 subcommands

#### Scenario: docs/cli/mcp.md exists with full reference
- **WHEN** a user reads docs/cli/mcp.md
- **THEN** each subcommand has argument tables, flag tables, and usage examples matching the actual CLI implementation

## MCP stdio Server OS Sandbox
