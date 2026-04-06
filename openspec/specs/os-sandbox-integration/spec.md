## ADDED Requirements

### Requirement: Exec tool sandbox integration
The exec tool SHALL apply `OSIsolator` to all 3 `exec.Command` call sites (`Run`, `RunWithPTY`, `StartBackground`) after command creation and before process start.

#### Scenario: Sandbox applied in Run
- **WHEN** `exec.Tool.Run()` is called with `Config.OSIsolator` set
- **THEN** `OSIsolator.Apply()` SHALL be called on the `exec.Cmd` before `cmd.Run()`

#### Scenario: Fail-closed rejects execution
- **WHEN** `Config.FailClosed` is true and `OSIsolator.Apply()` returns an error
- **THEN** the tool SHALL return `ErrSandboxRequired` without executing the command

#### Scenario: Fail-open logs warning
- **WHEN** `Config.FailClosed` is false and `OSIsolator.Apply()` returns an error
- **THEN** the tool SHALL log a warning and proceed with unsandboxed execution

### Requirement: MCP transport sandbox integration
The MCP `ServerConnection` SHALL apply `OSIsolator` to stdio transport `exec.Command` at transport creation time, with `MCPServerPolicy()` (network=allow, read-global, write-/tmp).

#### Scenario: Stdio transport sandboxed
- **WHEN** `createTransport()` is called for stdio transport with isolator set
- **THEN** `OSIsolator.Apply()` SHALL be called with `MCPServerPolicy()` before returning the transport

#### Scenario: Non-stdio transports not affected
- **WHEN** `createTransport()` is called for http or sse transport
- **THEN** `OSIsolator.Apply()` SHALL NOT be called

### Requirement: Skill script sandbox integration
The skill `Executor` SHALL apply `OSIsolator` to script execution `exec.Command` with `DefaultToolPolicy(workspacePath)`.

#### Scenario: Script execution sandboxed
- **WHEN** `executeScript()` is called with isolator set
- **THEN** `OSIsolator.Apply()` SHALL be called on the `exec.Cmd` before `cmd.Run()`

### Requirement: initOSSandbox logging
`initOSSandbox()` SHALL use `SandboxStatus` for structured logging. When the isolator is unavailable, log messages SHALL include the `Reason()` string and `PlatformCapabilities.Summary()`. Both fail-closed and fail-open paths SHALL log the reason.

#### Scenario: Sandbox disabled
- **WHEN** `sandbox.enabled` is false
- **THEN** `initOSSandbox()` SHALL return nil and no isolator is injected

#### Scenario: Fail-open logging with reason
- **WHEN** sandbox is enabled but isolator is unavailable with `failClosed=false`
- **THEN** log includes `reason` field with the isolator's `Reason()` value and `capabilities` field with `Summary()`

#### Scenario: Fail-closed logging with reason
- **WHEN** sandbox is enabled but isolator is unavailable with `failClosed=true`
- **THEN** log includes `reason` field with the isolator's `Reason()` value

### Requirement: Documentation accuracy
All code comments, doc comments, README, docs pages, and configuration references SHALL NOT claim Linux seccomp/Landlock enforcement when it is not implemented. Unimplemented features SHALL be marked as "planned" or "not yet enforced".

#### Scenario: Package doc comment
- **WHEN** reading the `sandbox/os` package documentation
- **THEN** it states Linux isolation is planned, not that it uses Landlock+seccomp

#### Scenario: Config field comments
- **WHEN** reading `SandboxConfig` field comments for Linux-specific behavior
- **THEN** they note Linux isolation is not yet enforced
