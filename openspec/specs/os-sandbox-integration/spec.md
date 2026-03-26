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

### Requirement: App wiring sandbox initialization
The app wiring SHALL create `OSIsolator` from `SandboxConfig` and inject it into exec tool, MCP manager, and skill executor via `initOSSandbox()` and `sandboxPolicy()`.

#### Scenario: Sandbox disabled
- **WHEN** `sandbox.enabled` is false
- **THEN** `initOSSandbox()` SHALL return nil and no isolator is injected
