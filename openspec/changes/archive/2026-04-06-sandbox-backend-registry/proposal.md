## Why

PR 1 (`linux-sandbox-hardening`) restored Linux build and added honest status reporting, but three critical gaps remained:

1. **Fail-closed only enforced in exec tool** — `skill executor` and `MCP manager` ignored `failClosed=true` and ran scripts/servers unsandboxed when sandbox was unavailable
2. **No backend selection mechanism** — `newPlatformIsolator()` was hardcoded; users could not configure or override which isolation backend to use
3. **Status output incomplete** — no way to inspect which backends are available on the current platform without reading source code

This change introduces a typed backend registry, propagates fail-closed across all process-launching paths, and exposes backend selection through config, CLI status, and TUI settings.

## What Changes

- Add `BackendMode` enum and `BackendCandidate` struct in `internal/sandbox/os/registry.go`
- Add `SelectBackend()` and `ListBackends()` functions for typed backend identity (no `Name()` string matching)
- Add `PlatformBackendCandidates()` shared helper used by both wiring and CLI to prevent drift
- Add stub isolators (`bwrapStub`, `nativeStub`) for planned backends with explicit "not yet implemented" reasons
- Add `Backend string` field to `SandboxConfig` (`auto`, `seatbelt`, `bwrap`, `native`, `none`)
- **BREAKING**: `config.Validate()` now rejects unknown `sandbox.backend` values at startup
- Add `SetFailClosed(bool)` method to `skill.Executor`, `skill.Registry`, `mcp.ServerConnection`, `mcp.ServerManager`
- Skill `executeScript()` and MCP `createTransport()` now reject execution when `failClosed=true` and sandbox unavailable
- `supervisor.New()`, `wiring_sandbox.go`, `wiring_knowledge.go`, `wiring_mcp.go` consume the registry via `SelectBackend()`
- `backend=none` is treated as explicit opt-out — fail-closed does not apply, all paths run unsandboxed
- `lango sandbox status` adds Backend Availability section, Backend resolution line, and `none` opt-out display
- `lango sandbox test` honors the configured backend (previously always used `NewOSIsolator()`)
- TUI settings form gains `os_sandbox_backend` select field (auto/seatbelt/bwrap/native/none)
- Documentation updated: `README.md`, `docs/configuration.md`, `docs/cli/sandbox.md`

## Capabilities

### New Capabilities

- `sandbox-backend-registry`: Typed backend selection with `BackendMode`/`BackendCandidate`/`SelectBackend`/`ListBackends`/`PlatformBackendCandidates` and explicit `bwrap`/`native` stub registration
- `sandbox-fail-closed-enforcement`: Skill executor and MCP connection respect `failClosed=true` by rejecting execution when sandbox is unavailable, matching the existing exec tool behavior

### Modified Capabilities

- `os-sandbox-core`: `OSIsolator` interface unchanged; new `BackendMode`/`BackendCandidate`/`BackendInfo`/`SelectBackend`/`ListBackends`/`PlatformBackendCandidates` symbols added; auto-fallback noop preserves aggregated candidate reasons
- `os-sandbox-cli`: `sandbox status` adds Backend Availability section + opt-out display; `sandbox test` honors `cfg.Sandbox.Backend`
- `os-sandbox-integration`: `initOSSandbox()`, `supervisor.New()`, `wiring_knowledge.go`, `wiring_mcp.go` use the registry; `backend=none` short-circuits sandbox wiring

## Impact

- **Config validation breaking change**: invalid `sandbox.backend` values now block startup. Users must use one of `auto`/`seatbelt`/`bwrap`/`native`/`none` (default `auto`)
- **Behavioral change**: skill scripts and MCP stdio servers now respect `sandbox.failClosed` — any deployment with `failClosed=true` and an unavailable sandbox will start failing those paths (previously silently ran unsandboxed)
- **`backend=none` semantics**: explicit opt-out — equivalent to `enabled=false` for execution but still surfaced in status output
- 14 files modified, 3 new (`registry.go`, `registry_test.go`, ~~existing~~)
- No external dependency changes
