## Why

PR 1~3 of the Linux sandbox roadmap restored the build, added a backend registry, and shipped the first real Linux isolation backend (`bwrap`). What remains is operational hardening: the sandboxed children produced by exec / skill / MCP could still read the lango control-plane (`~/.lango`), there was no controlled bypass for commands the user explicitly trusts, fail-open fallback was silent, and the audit log had no record of sandbox decisions. This change closes those four gaps in one PR so the sandbox layer is honest about what it does, what it skips, and what it lets through.

## What Changes

- `NormalizePaths` (loader.go) now also normalizes `Sandbox.WorkspacePath`, `Sandbox.AllowedWritePaths` (slice), and `Sandbox.OS.SeatbeltCustomProfile`. New `normalizePathSlice` helper.
- **BREAKING (internal API)**: `DefaultToolPolicy`, `StrictToolPolicy`, and `MCPServerPolicy` all gain a `dataRoot` parameter. Pass `cfg.DataRoot` from the wiring; pass `""` to skip the control-plane mask in unit tests.
- All three policies now deny the workspace's `.git` directory as a baseline (was previously a `StrictToolPolicy`-only feature). All three deny `dataRoot` when non-empty so sandboxed children cannot read or write `~/.lango/*`.
- `Sandbox.AllowedWritePaths` is now actually wired: `supervisor.go` appends each entry to the exec tool's policy WritePaths.
- New config field `Sandbox.ExcludedCommands []string`. When the user command's first whitespace-separated token (basename) matches an entry, `applySandbox` returns immediately without applying the isolator. Matching is performed against the user command (NOT `cmd.Args[0]`, which is always `"sh"` because exec.Tool wraps everything in `sh -c`). Chained commands like `cd /tmp && git status` do NOT match — only direct invocations.
- New event `SandboxDecisionEvent` in eventbus, with helper `PublishSandboxDecision`. Audit recorder subscribes and writes one `AuditLog` row per decision (action="sandbox_decision"). New ent enum value via `go generate`.
- All three sandbox apply call sites now publish `SandboxDecisionEvent`: `exec.Tool.applySandbox`, `skill.Executor.executeScript`, `mcp.ServerConnection.createTransport`. SessionKey is derived from runtime ctx via `session.SessionKeyFromContext`; for MCP it is empty (process-level lifecycle).
- Fail-open fallback in `exec.Tool` now also prints a one-shot stderr warning via `sync.Once` so the user notices that subsequent commands are unsandboxed. The full per-command audit trail is in `lango sandbox status` instead of repeated stderr noise.
- `lango sandbox status` gains a "Recent Sandbox Decisions" section: last 10 audit rows with timestamp, session-prefix, decision, backend, target, optional reason. New `--session <prefix>` flag for session filtering. Graceful degradation: nil bootLoader / DB locked / signed-out all silently skip the section so the diagnostic remains usable as a pure sandbox-layer inspection tool.
- `NewSandboxCmd` signature gains a `BootLoader` parameter (optional). Wired in `cmd/lango/main.go`.
- TUI OS Sandbox form gains `os_sandbox_excluded_commands` InputText field (comma-separated). `state_update.go` maps it to `cfg.Sandbox.ExcludedCommands` via `splitCSV`.
- Wiring: `initSkills(cfg, baseTools, bus)` and `initMCP(cfg, bus)` gain a bus parameter. `Registry.SetEventBus` and `ServerManager.SetEventBus` pass-throughs. `Supervisor.SetEventBus` is called from `app.go` after `populateAppFields`.
- README, `docs/configuration.md`, `docs/cli/sandbox.md`, `prompts/TOOL_USAGE.md`, `prompts/SAFETY.md` updated to document control-plane masking, ExcludedCommands semantics, fail-open warning, and the Recent Decisions section.

## Capabilities

### New Capabilities

- `sandbox-exception-policy`: Defines the `ExcludedCommands` bypass model, the `SandboxDecisionEvent` schema, the fail-open user-visible warning contract, and the rule that all three sandbox call sites must publish decision events.

### Modified Capabilities

- `os-sandbox-core`: Policy helpers gain `dataRoot` parameter; control-plane and `.git` denied as a baseline.
- `os-sandbox-cli`: `lango sandbox status` adds Recent Sandbox Decisions section with `--session` flag.
- `os-sandbox-integration`: Supervisor wires `cfg.DataRoot` and `AllowedWritePaths`; skill executor and MCP transport carry dataRoot through their wiring; all three sites publish `SandboxDecisionEvent`.
- `mcp-integration`: `MCPServerPolicy` takes `dataRoot`; `ServerConnection` carries dataRoot and bus; `createTransport` publishes a process-level `SandboxDecisionEvent` with empty SessionKey.
- `config-system`: `NormalizePaths` extended to cover sandbox path fields.

## Impact

- **Affected code**: `internal/config/loader.go`, `internal/config/types_sandbox.go`, `internal/sandbox/os/policy.go` and tests, `internal/sandbox/os/bwrap_args_test.go`, `internal/sandbox/os/bwrap_linux_test.go`, `internal/supervisor/supervisor.go`, `internal/skill/{executor,registry}.go` and tests, `internal/mcp/{connection,manager}.go` and tests, `internal/tools/exec/exec.go` and tests, `internal/eventbus/events.go`, `internal/observability/audit/recorder.go`, `internal/ent/schema/audit_log.go` (+ `go generate`), `internal/cli/sandbox/sandbox.go` (+ new test file), `internal/cli/settings/forms_sandbox.go`, `internal/cli/tuicore/state_update.go`, `internal/app/{app,modules,wiring_knowledge,wiring_mcp}.go`, `cmd/lango/main.go`.
- **Affected docs**: `README.md`, `docs/configuration.md`, `docs/cli/sandbox.md`, `prompts/TOOL_USAGE.md`, `prompts/SAFETY.md`.
- **Schema migration**: ent enum `audit_log.action` gains `"sandbox_decision"` value. Backward compatible (enum addition only).
- **No external API breakage**: only internal package signatures change. The wired CLI surface (`lango sandbox status`, settings form) gains fields and a flag without removing anything.
- **Risks**: `.git` denial as a baseline may break agent-driven git workflows when bubblewrap or Seatbelt is active. Mitigation: users can add `git` to `sandbox.excludedCommands` to bypass for that specific command.
- **Out of scope (PR 5)**: Native Landlock+seccomp backend, file-level deny via `--ro-bind /dev/null <file>`, symlink chain resolution before bwrap arg compile, glob/path semantics normalization, per-tool policy override.
