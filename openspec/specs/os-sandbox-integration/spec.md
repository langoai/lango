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

### Requirement: initOSSandbox uses backend registry
`initOSSandbox(cfg *config.Config)` SHALL parse `cfg.Sandbox.Backend`, build candidates via `PlatformBackendCandidates()`, and call `SelectBackend(mode, candidates)` instead of `NewOSIsolator()` directly. When `cfg.Sandbox.Enabled=false` OR `mode == BackendNone`, the function SHALL return nil to signal "no sandbox wiring required". Backend validity is enforced by `config.Validate()`; the parse error in this function is therefore unreachable and discarded. When the isolator is unavailable, log messages SHALL include the `Reason()` string, the backend label, and `PlatformCapabilities.Summary()`.

#### Scenario: Sandbox disabled
- **WHEN** `sandbox.enabled` is false
- **THEN** `initOSSandbox()` SHALL return nil and no isolator is injected

#### Scenario: backend=none returns nil
- **WHEN** `cfg.Sandbox.Enabled=true` and `cfg.Sandbox.Backend="none"`
- **THEN** `initOSSandbox()` returns nil and logs `"OS sandbox disabled via backend=none (explicit opt-out)"`

#### Scenario: Auto selection logged with backend label
- **WHEN** `initOSSandbox()` selects an available backend via auto mode
- **THEN** the info log includes both `isolator` (e.g., "seatbelt") and `backend` (e.g., "auto") fields

#### Scenario: Fail-open logging with reason
- **WHEN** sandbox is enabled but isolator is unavailable with `failClosed=false`
- **THEN** log includes `reason` field with the isolator's `Reason()` value and `capabilities` field with `Summary()`

#### Scenario: Fail-closed logging with reason
- **WHEN** sandbox is enabled but isolator is unavailable with `failClosed=true`
- **THEN** log includes `reason` field with the isolator's `Reason()` value

### Requirement: supervisor consumes backend registry
`supervisor.New()` SHALL use `ParseBackendMode + SelectBackend(mode, PlatformBackendCandidates())` to build the exec tool's `OSIsolator` and `FailClosed`. When `mode == BackendNone`, supervisor SHALL skip sandbox wiring entirely (no `OSIsolator`, no `FailClosed=true`) so that exec commands run unsandboxed without rejection.

#### Scenario: backend=none skips exec tool sandbox wiring
- **WHEN** `cfg.Sandbox.Enabled=true`, `cfg.Sandbox.Backend="none"`, `cfg.Sandbox.FailClosed=true`
- **THEN** `supervisor.New()` does NOT set `execConfig.OSIsolator` or `execConfig.FailClosed`, and exec commands run without rejection

#### Scenario: backend selection drives exec isolator
- **WHEN** `cfg.Sandbox.Backend="seatbelt"` and Seatbelt is available
- **THEN** `execConfig.OSIsolator` is the SeatbeltIsolator returned by `SelectBackend`

### Requirement: Skill and MCP wiring propagate fail-closed
`wiring_knowledge.go` and `wiring_mcp.go` SHALL call `SetFailClosed(cfg.Sandbox.FailClosed)` on `skill.Registry` and `mcp.ServerManager` respectively, after `initOSSandbox()` returns a non-nil isolator. When `initOSSandbox()` returns nil (disabled or backend=none), these wiring paths SHALL skip both `SetOSIsolator` and `SetFailClosed`.

#### Scenario: Fail-closed propagated to skill registry
- **WHEN** sandbox is enabled with `failClosed=true` and a backend is wired
- **THEN** `wiring_knowledge.go` calls `registry.SetFailClosed(true)` after `SetOSIsolator`

#### Scenario: backend=none bypasses skill fail-closed
- **WHEN** `backend=none` and `failClosed=true`
- **THEN** `wiring_knowledge.go` skips both `SetOSIsolator` and `SetFailClosed`, allowing scripts to run unsandboxed

### Requirement: Config validates sandbox.backend at startup
`config.Validate()` SHALL call `sandboxos.ParseBackendMode(cfg.Sandbox.Backend)` and append a validation error when the value is not one of `auto`, `seatbelt`, `bwrap`, `native`, or `none`. The error message SHALL list the valid values.

#### Scenario: Typo rejected at startup
- **WHEN** config sets `sandbox.backend: "seatbeltt"`
- **THEN** `config.Validate()` returns an error containing `"sandbox.backend"` and `"must be auto, seatbelt, bwrap, native, or none"`

#### Scenario: Empty string accepted (defaults to auto)
- **WHEN** `sandbox.backend` is empty
- **THEN** `config.Validate()` does not return an error and the runtime treats it as `auto`

### Requirement: Documentation accuracy
All code comments, doc comments, README, docs pages, and configuration references SHALL NOT claim Linux seccomp/Landlock enforcement when it is not implemented. Unimplemented features SHALL be marked as "planned" or "not yet enforced".

#### Scenario: Package doc comment
- **WHEN** reading the `sandbox/os` package documentation
- **THEN** it states Linux isolation is planned, not that it uses Landlock+seccomp

#### Scenario: Config field comments
- **WHEN** reading `SandboxConfig` field comments for Linux-specific behavior
- **THEN** they note Linux isolation is not yet enforced
