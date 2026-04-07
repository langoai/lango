## MODIFIED Requirements

### Requirement: initOSSandbox uses backend registry
`initOSSandbox(cfg *config.Config)` SHALL parse `cfg.Sandbox.Backend`, build candidates via `PlatformBackendCandidates()`, and call `SelectBackend(mode, candidates)` instead of `NewOSIsolator()` directly. When `cfg.Sandbox.Enabled=false` OR `mode == BackendNone`, the function SHALL return nil to signal "no sandbox wiring required".

Backend validity is enforced by `config.Validate()`; the parse error in this function is therefore unreachable and discarded.

#### Scenario: backend=none returns nil
- **WHEN** `cfg.Sandbox.Enabled=true` and `cfg.Sandbox.Backend="none"`
- **THEN** `initOSSandbox()` returns nil and logs `"OS sandbox disabled via backend=none (explicit opt-out)"`

#### Scenario: Auto selection logged with backend label
- **WHEN** `initOSSandbox()` selects an available backend via auto mode
- **THEN** the info log includes both `isolator` (e.g., "seatbelt") and `backend` (e.g., "auto") fields

### Requirement: supervisor consumes backend registry
`supervisor.New()` SHALL use `ParseBackendMode + SelectBackend(mode, PlatformBackendCandidates())` to build the exec tool's `OSIsolator` and `FailClosed`. When `mode == BackendNone`, supervisor SHALL skip sandbox wiring entirely (no `OSIsolator`, no `FailClosed=true`) so that exec commands run unsandboxed without rejection.

#### Scenario: backend=none skips exec tool sandbox wiring
- **WHEN** `cfg.Sandbox.Enabled=true`, `cfg.Sandbox.Backend="none"`, `cfg.Sandbox.FailClosed=true`
- **THEN** `supervisor.New()` does NOT set `execConfig.OSIsolator` or `execConfig.FailClosed`, and exec commands run without rejection

#### Scenario: backend selection drives exec isolator
- **WHEN** `cfg.Sandbox.Backend="seatbelt"` and Seatbelt is available
- **THEN** `execConfig.OSIsolator` is the SeatbeltIsolator returned by `SelectBackend`

### Requirement: skill and MCP wiring propagate fail-closed
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
