## MODIFIED Requirements

### Requirement: MCP transport sandbox integration
The MCP `ServerConnection` SHALL apply `OSIsolator` to stdio transport `exec.Command` at transport creation time, with `MCPServerPolicy(workspacePath, dataRoot)` (network=allow, read-global, write-/tmp, ancestor `.git` and dataRoot denied).

`ServerConnection` SHALL carry a `workspacePath string` field AND a `dataRoot string` field, both set via `SetOSIsolator(iso, workspacePath, dataRoot)`, and an optional `bus *eventbus.Bus` field set via `SetEventBus(bus)`. `createTransport()` SHALL pass both `sc.workspacePath` and `sc.dataRoot` to `MCPServerPolicy`. When `sc.workspacePath` is non-empty, `createTransport()` SHALL ALSO set `cmd.Dir = sc.workspacePath` so the spawned MCP child runs with cwd inside the user's workspace — policy discovery (walk-up to `.git`) and execution share the same git context. An empty `workspacePath` SHALL leave `cmd.Dir` unset, preserving legacy behavior (supervisor cwd inherited).

`createTransport()` SHALL publish a `SandboxDecisionEvent` with `Source="mcp"` and `Command=sc.name` for every decision: `applied`, `skipped`, `rejected`, AND for the `failClosed && isolator==nil` rejection path. The `SessionKey` field SHALL be empty because MCP server lifecycle is process-level, not session-bound.

`mcp.ServerManager` SHALL gain a `workspacePath string` field alongside the existing `dataRoot` field, set via `SetOSIsolator(iso, workspacePath, dataRoot)`. `ConnectAll` SHALL pass the manager's `workspacePath`, `dataRoot`, and `bus` to each newly-created connection.

#### Scenario: Stdio transport sandboxed with workspacePath and dataRoot
- **WHEN** `createTransport()` is called for stdio transport with isolator, workspacePath, and dataRoot all set
- **THEN** `OSIsolator.Apply()` SHALL be called with `MCPServerPolicy(sc.workspacePath, sc.dataRoot)` before returning the transport
- **AND** `cmd.Dir` SHALL equal `sc.workspacePath`
- **AND** a `SandboxDecisionEvent{Source:"mcp", Decision:"applied"}` SHALL be published with empty SessionKey

#### Scenario: Empty workspacePath leaves cmd.Dir unset
- **WHEN** `createTransport()` is called for stdio transport AND `sc.workspacePath` is empty
- **THEN** `cmd.Dir` SHALL NOT be set by `createTransport()` (preserves legacy behavior — Go's exec.Cmd inherits supervisor cwd)
- **AND** `MCPServerPolicy("", sc.dataRoot)` SHALL be called (policy silently skips the walk-up deny)

#### Scenario: Non-stdio transports not affected
- **WHEN** `createTransport()` is called for http or sse transport
- **THEN** `OSIsolator.Apply()` SHALL NOT be called and no `SandboxDecisionEvent` SHALL be published
