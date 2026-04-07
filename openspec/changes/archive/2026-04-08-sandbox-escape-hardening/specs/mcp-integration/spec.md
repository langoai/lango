## MODIFIED Requirements

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
