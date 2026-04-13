## ADDED Requirements

### Requirement: Skill executor rejects unsandboxed execution under fail-closed
`skill.Executor` SHALL provide a `SetFailClosed(bool)` method. When `failClosed=true` and either (a) no isolator is configured or (b) `isolator.Apply()` returns an error, `executeScript()` SHALL return `ErrSandboxRequired` instead of running the script unsandboxed.

#### Scenario: Nil isolator with fail-closed
- **WHEN** `failClosed=true` and no isolator is set
- **THEN** `executeScript()` returns an error wrapping `ErrSandboxRequired` with message `"no OS isolator configured for skill script"`

#### Scenario: Apply error with fail-closed
- **WHEN** `failClosed=true` and the isolator's `Apply()` returns an error
- **THEN** `executeScript()` returns an error wrapping `ErrSandboxRequired` and the original Apply error

#### Scenario: Apply error in fail-open mode still proceeds
- **WHEN** `failClosed=false` and the isolator's `Apply()` returns an error
- **THEN** `executeScript()` logs a warning and runs the script unsandboxed (existing behavior)

### Requirement: Skill registry propagates fail-closed to executor
`skill.Registry` SHALL provide a `SetFailClosed(bool)` method that delegates to its `Executor`.

#### Scenario: Registry SetFailClosed delegates
- **WHEN** `registry.SetFailClosed(true)` is called
- **THEN** the underlying executor's `failClosed` field becomes true

### Requirement: MCP connection rejects stdio transport under fail-closed
`mcp.ServerConnection` SHALL provide a `SetFailClosed(bool)` method. When `failClosed=true` and either (a) no isolator is configured or (b) `isolator.Apply()` returns an error, `createTransport()` for stdio servers SHALL return an error wrapping `ErrSandboxRequired`. HTTP/SSE transports SHALL NOT be affected.

#### Scenario: Stdio with nil isolator and fail-closed
- **WHEN** server transport is `stdio`, `failClosed=true`, and no isolator is set
- **THEN** `createTransport()` returns an error wrapping `ErrSandboxRequired` with the server name

#### Scenario: HTTP transport unaffected
- **WHEN** server transport is `http` and `failClosed=true`
- **THEN** `createTransport()` succeeds regardless of isolator state

### Requirement: MCP manager propagates fail-closed to current and future connections
`mcp.ServerManager` SHALL provide a `SetFailClosed(bool)` method that:
- Stores the value on the manager
- Propagates to all currently registered connections
- Applies to future connections created during `ConnectAll()` / `Connect()`

#### Scenario: Future connection inherits fail-closed
- **WHEN** `mgr.SetFailClosed(true)` is called BEFORE `ConnectAll()`
- **THEN** every connection created during `ConnectAll()` has `failClosed=true`

#### Scenario: Existing connection updated
- **WHEN** `mgr.SetFailClosed(true)` is called AFTER connections are registered
- **THEN** each existing connection's `failClosed` field becomes true
