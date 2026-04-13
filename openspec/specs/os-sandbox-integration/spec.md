## Purpose

Integration layer that wires the `OSIsolator` and its `Policy` helpers into the exec tool, skill executor, MCP transport, supervisor, CLI status, and config validator. This capability covers how sandbox is applied, how failures are reported (fail-open vs fail-closed), and how audit events surface through the event bus for every Apply() call site.
## Requirements
### Requirement: Exec tool sandbox integration
The exec tool SHALL apply `OSIsolator` to all 3 `exec.Command` call sites (`Run`, `RunWithPTY`, `StartBackground`) after command creation and before process start.

`Tool.applySandbox(ctx, cmd, userCommand string)` SHALL accept the raw user command string as a third parameter so the bypass matcher and audit publisher can see the pre-`sh -c`, pre-secret-resolution command. Each of the three call sites SHALL pass the original `command` argument (not the resolved string).

When `Config.ExcludedCommands` is non-empty and the basename of the user command's first whitespace-separated token matches an entry, `applySandbox` SHALL return `nil` immediately without calling `OSIsolator.Apply` and SHALL publish a `SandboxDecisionEvent` with `Decision="excluded"` and the matched `Pattern`. The matcher SHALL NOT use `cmd.Args[0]` (which is always `"sh"` because exec.Tool wraps commands in `sh -c`).

When `OSIsolator.Apply` succeeds, the tool SHALL publish a `SandboxDecisionEvent` with `Decision="applied"` and `Backend` set from `Isolator.Name()`. When it returns an error and `FailClosed=false`, the tool SHALL publish `Decision="skipped"` with the error reason, log a warning, AND emit a one-shot stderr message via `sync.Once` so the user notices that subsequent commands are running unsandboxed. When it returns an error and `FailClosed=true`, the tool SHALL publish `Decision="rejected"` and return `ErrSandboxRequired`.

The `SessionKey` field on every published event SHALL be derived from the runtime context via `session.SessionKeyFromContext(ctx)`. The exec tool SHALL NOT store a session key on its `Config` or `Tool` struct.

#### Scenario: Sandbox applied in Run
- **WHEN** `exec.Tool.Run()` is called with `Config.OSIsolator` set
- **THEN** `OSIsolator.Apply()` SHALL be called on the `exec.Cmd` before `cmd.Run()`
- **AND** a `SandboxDecisionEvent{Decision:"applied", Source:"exec"}` SHALL be published

#### Scenario: Fail-closed rejects execution
- **WHEN** `Config.FailClosed` is true and `OSIsolator.Apply()` returns an error
- **THEN** the tool SHALL return `ErrSandboxRequired` without executing the command
- **AND** a `SandboxDecisionEvent{Decision:"rejected"}` SHALL be published

#### Scenario: Fail-open logs warning + one-shot stderr
- **WHEN** `Config.FailClosed` is false and `OSIsolator.Apply()` returns an error
- **THEN** the tool SHALL log a warning, emit one stderr line of the form `lango: WARNING — sandbox fallback active (reason: ...); commands run unsandboxed`, publish a `SandboxDecisionEvent{Decision:"skipped"}`, and proceed with unsandboxed execution
- **AND** a subsequent fallback in the same process SHALL NOT emit a duplicate stderr line

#### Scenario: Excluded command bypasses sandbox
- **WHEN** `Config.ExcludedCommands` contains `"git"` and the user command is `"git status"`
- **THEN** `OSIsolator.Apply` SHALL NOT be called and a `SandboxDecisionEvent{Decision:"excluded", Pattern:"git"}` SHALL be published

#### Scenario: Excluded does not match sh wrapper
- **WHEN** `Config.ExcludedCommands` contains `"sh"` and the user command is `"echo hello"`
- **THEN** the sandbox SHALL be applied normally (the matcher consumes the user command, not `cmd.Args[0]`)

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

### Requirement: Skill script sandbox integration
The skill `Executor` SHALL apply `OSIsolator` to script execution `exec.Command` with `DefaultToolPolicy(workspacePath, dataRoot)`.

`Executor` SHALL carry a `dataRoot string` field set via `SetOSIsolator(iso, workspacePath, dataRoot)` and an optional `bus *eventbus.Bus` field set via `SetEventBus(bus)`. `executeScript(ctx, ...)` SHALL pass `e.dataRoot` to `DefaultToolPolicy`.

`executeScript` SHALL publish a `SandboxDecisionEvent` with `Source="skill"` and `Command=skill.Name` for every decision branch: `applied` (after `Apply` succeeds), `skipped` (after `Apply` fails with `failClosed=false`), `rejected` (after `Apply` fails with `failClosed=true`), AND for the `failClosed && isolator==nil` rejection path. The `SessionKey` field SHALL be derived from the runtime context via `session.SessionKeyFromContext(ctx)`.

`skill.Registry` SHALL provide `SetOSIsolator(iso, workspacePath, dataRoot)` and `SetEventBus(bus)` pass-throughs that forward to the underlying `Executor`. The wiring layer (`wiring_knowledge.go:initSkills`) SHALL call these on the registry, not on the executor directly.

#### Scenario: Script execution sandboxed with dataRoot and bus
- **WHEN** `executeScript()` is called with isolator, dataRoot, and bus all set
- **THEN** `OSIsolator.Apply()` SHALL be called with `DefaultToolPolicy(e.workspacePath, e.dataRoot)` before `cmd.Run()`
- **AND** a `SandboxDecisionEvent{Source:"skill", Decision:"applied", Command:skill.Name}` SHALL be published

#### Scenario: Skill SessionKey derived from ctx
- **WHEN** `executeScript()` is called with a context carrying a session key (via `session.WithSessionKey(ctx, "test")`)
- **THEN** the published `SandboxDecisionEvent.SessionKey` SHALL equal `"test"`

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

When wiring the exec tool's policy, supervisor SHALL call `DefaultToolPolicy(workDir, cfg.DataRoot)` and SHALL append every entry of `cfg.Sandbox.AllowedWritePaths` to `policy.Filesystem.WritePaths` so the previously-dead config field becomes effective.

`supervisor.New()` SHALL also set `execConfig.ExcludedCommands` from `cfg.Sandbox.ExcludedCommands` (defensively copied).

`Supervisor` SHALL provide a `SetEventBus(*eventbus.Bus)` method that forwards to the exec tool. `app.go` post-build wiring (B1a) SHALL resolve the supervisor from the resolver and call `SetEventBus(bus)` so SandboxDecisionEvent records flow into audit.

#### Scenario: backend=none skips exec tool sandbox wiring
- **WHEN** `cfg.Sandbox.Enabled=true`, `cfg.Sandbox.Backend="none"`, `cfg.Sandbox.FailClosed=true`
- **THEN** `supervisor.New()` does NOT set `execConfig.OSIsolator` or `execConfig.FailClosed`, and exec commands run without rejection

#### Scenario: backend selection drives exec isolator
- **WHEN** `cfg.Sandbox.Backend="seatbelt"` and Seatbelt is available
- **THEN** `execConfig.OSIsolator` is the SeatbeltIsolator returned by `SelectBackend`

#### Scenario: AllowedWritePaths appended to exec policy
- **WHEN** `cfg.Sandbox.AllowedWritePaths = ["/tmp/scratch", "/var/cache/app"]` and the supervisor builds the exec tool
- **THEN** `execConfig.SandboxPolicy.Filesystem.WritePaths` SHALL contain both entries in addition to `workDir` and `/tmp`

#### Scenario: ExcludedCommands forwarded to exec tool
- **WHEN** `cfg.Sandbox.ExcludedCommands = ["git"]`
- **THEN** the constructed `exec.Tool.Config.ExcludedCommands` SHALL equal `["git"]`

#### Scenario: SetEventBus forwards to exec tool
- **WHEN** `Supervisor.SetEventBus(bus)` is called after construction
- **THEN** subsequent `SandboxDecisionEvent` publishes from the exec tool SHALL be delivered on `bus`

### Requirement: Skill and MCP wiring propagate fail-closed and event bus
`wiring_knowledge.go` and `wiring_mcp.go` SHALL call `SetFailClosed(cfg.Sandbox.FailClosed)` on `skill.Registry` and `mcp.ServerManager` respectively, after `initOSSandbox()` returns a non-nil isolator. When `initOSSandbox()` returns nil (disabled or backend=none), these wiring paths SHALL skip both `SetOSIsolator` and `SetFailClosed`.

`initSkills` SHALL accept a `bus *eventbus.Bus` parameter and call `registry.SetEventBus(bus)` when `bus` is non-nil. Similarly, `initMCP` SHALL accept a `bus *eventbus.Bus` parameter and call `mgr.SetEventBus(bus)`.

The skill/MCP wiring SHALL also propagate `cfg.DataRoot` so the underlying executor and ServerConnection get the control-plane mask: `registry.SetOSIsolator(iso, workDir, cfg.DataRoot)` and `mgr.SetOSIsolator(iso, cfg.DataRoot)`.

#### Scenario: Fail-closed propagated to skill registry
- **WHEN** sandbox is enabled with `failClosed=true` and a backend is wired
- **THEN** `wiring_knowledge.go` calls `registry.SetFailClosed(true)` after `SetOSIsolator`

#### Scenario: backend=none bypasses skill fail-closed
- **WHEN** `backend=none` and `failClosed=true`
- **THEN** `wiring_knowledge.go` skips both `SetOSIsolator` and `SetFailClosed`, allowing scripts to run unsandboxed

#### Scenario: Event bus propagated to skill registry
- **WHEN** `initSkills` is called with a non-nil bus
- **THEN** `registry.SetEventBus(bus)` SHALL be called so SandboxDecisionEvent records from skill scripts reach audit

#### Scenario: Event bus propagated to MCP manager
- **WHEN** `initMCP` is called with a non-nil bus
- **THEN** `mgr.SetEventBus(bus)` SHALL be called and every newly-connected `ServerConnection` SHALL receive that bus

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

