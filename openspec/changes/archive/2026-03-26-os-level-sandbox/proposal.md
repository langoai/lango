## Why

Lango's sandbox is currently P2P-only (`internal/sandbox/`, subprocess/container isolation for remote tool calls). Industry leaders — Claude Code, Cursor, Codex CLI — apply OS-level kernel sandboxing (Seatbelt on macOS, Landlock+seccomp on Linux) to **all** shell child processes. Local exec tools, MCP stdio servers, and skill scripts currently spawn child processes with full parent process privileges — no filesystem scoping, no network restriction, no syscall filtering.

## What Changes

- New `internal/sandbox/os/` package: `OSIsolator` interface, Seatbelt profile generation (macOS), Landlock+seccomp stubs (Linux), platform capability probe, cross-platform build-tag stubs
- Exec tool (`internal/tools/exec/exec.go`): `applySandbox()` inserted at 3 `exec.Command` sites (Run, RunWithPTY, StartBackground) with fail-open/fail-closed policy
- MCP transport (`internal/mcp/connection.go`): `SetOSIsolator()` applied at stdio transport `exec.Command` creation
- Skill executor (`internal/skill/executor.go`): `SetOSIsolator()` applied before script `exec.Command` run
- App wiring (`internal/app/wiring_sandbox.go`): `initOSSandbox()` + `sandboxPolicy()` for DI into exec/MCP/skill
- CLI (`internal/cli/sandbox/sandbox.go`): `lango sandbox status` and `lango sandbox test`
- Config (`internal/config/types_sandbox.go`): `SandboxConfig` with `enabled`, `failClosed`, `networkMode`, `allowedNetworkIPs` (macOS only), `allowedWritePaths`

## Capabilities

### New Capabilities
- `os-sandbox-core`: OSIsolator interface, Policy types, Seatbelt profile generation, platform capability probe, build-tag stubs
- `os-sandbox-integration`: Exec tool, MCP transport, skill script sandbox wiring at exec.Command sites
- `os-sandbox-cli`: `lango sandbox status` and `lango sandbox test` CLI commands

### Modified Capabilities
- `tool-sandbox`: Extended from P2P-only to include OS-level kernel primitives for general tool execution
- `tool-exec`: Exec tool now supports optional OS-level sandbox via Config.OSIsolator
- `mcp-integration`: MCP ServerConnection now supports optional OS-level sandbox for stdio transports

## Impact

- **Code**: `internal/sandbox/os/` (new, 13 files), `internal/tools/exec/exec.go`, `internal/mcp/connection.go`, `internal/skill/executor.go`, `internal/app/wiring_sandbox.go`, `internal/cli/sandbox/`, `internal/config/types_sandbox.go`, `cmd/lango/main.go`
- **Config**: New `sandbox.*` section independent of existing `p2p.toolIsolation`
- **Platform**: macOS Seatbelt (sandbox-exec), Linux Landlock+seccomp stubs. No CGO required.
- **Dependencies**: No new external dependencies (uses `golang.org/x/sys/unix` already present)
