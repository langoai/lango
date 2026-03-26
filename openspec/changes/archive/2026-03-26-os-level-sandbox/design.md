## Context

Lango's existing sandbox (`internal/sandbox/`) provides subprocess/container isolation for P2P remote tool calls only. Industry standard (Claude Code, Cursor, Codex CLI) applies OS-level kernel sandboxing to all child processes spawned by the agent. This design extends sandbox coverage to exec tools, MCP stdio servers, and skill scripts by applying OS primitives at `exec.Command` call sites.

Key constraint: In-process tools (fs, browser, knowledge) are bound to live dependencies (Supervisor, SessionManager, EntStore) via closures — kernel isolation is not applicable. Only child processes created via `exec.Command` are sandboxable.

## Goals / Non-Goals

**Goals:**
- Apply OS-level kernel sandbox (Seatbelt/Landlock+seccomp) to child processes at `exec.Command` call sites
- Support fail-open (default) and fail-closed modes
- Platform-aware: macOS Seatbelt with IP allowlist, Linux Landlock+seccomp with full network deny only
- Independent of existing `p2p.toolIsolation` config — no migration needed
- CLI observability: `lango sandbox status/test`

**Non-Goals:**
- WithSandbox middleware (isolation is at exec.Command, not handler level)
- Worker ToolRegistry extension (subprocess delegation of local tools)
- Domain-based network allowlist on Linux (seccomp sees IP only)
- In-process tool kernel isolation (fs_write, browser_*, knowledge_*)
- Prefix-based tool classification

## Decisions

### 1. Apply at exec.Command sites, not as middleware
**Choice**: Insert `OSIsolator.Apply(ctx, cmd, policy)` after `exec.Command()` and before `cmd.Run()`/`cmd.Start()`.
**Rationale**: Tool handlers for exec/MCP/skill create child processes internally. A middleware would need to intercept the child process creation, but the handler itself owns the `exec.Cmd`. Applying at the call site is how Claude Code and Cursor actually implement it.
**Alternative rejected**: WithSandbox middleware — requires serializing tool state to subprocess worker, which is infeasible for stateful tools.

### 2. OSIsolator interface with platform build tags
**Choice**: Single `OSIsolator` interface with `seatbelt_darwin.go`, `landlock_linux.go`, `seccomp_linux.go`, `*_stub.go` stubs.
**Rationale**: Go build tags cleanly separate platform-specific code. The noopIsolator fallback ensures compilation on all platforms. Factory `NewOSIsolator()` returns best available.

### 3. Config field `allowedNetworkIPs` not `allowedDomains`
**Choice**: IP-based field name reflecting actual Seatbelt `remote ip-literal` capability.
**Rationale**: Seatbelt operates on resolved IPs, not domain names. Naming it `allowedDomains` would overstate the guarantee. DNS resolution happens at profile generation time; CDN/IP changes require profile regeneration.

### 4. Independent `sandbox.*` config, not extending `p2p.toolIsolation`
**Choice**: New top-level `sandbox.*` config section.
**Rationale**: Different isolation mechanism (OS kernel primitives vs subprocess/container delegation), different targets (exec.Command child processes vs P2P tool handlers), different lifecycles. TUI/README/OpenSpec for `p2p.toolIsolation` remain unchanged.

### 5. Seatbelt profile via text/template, not string concatenation
**Choice**: `text/template` for .sb profile generation with path sanitization.
**Rationale**: Prevents S-expression injection via crafted paths. All paths validated against character allowlist before embedding.

## Risks / Trade-offs

- **[Seatbelt deprecated since macOS 10.15]** → Claude Code/Cursor also use it; still functional through macOS 15. `Available()` probe detects removal.
- **[PTY + Seatbelt compatibility]** → `pty.Start(cmd)` with sandbox-exec wrapped cmd needs verification. Fallback: fail-open degrades gracefully.
- **[Linux Landlock kernel version]** → Requires kernel 5.13+. ABI probe detects availability; falls back to seccomp-only or noop.
- **[Silent degradation in fail-open mode]** → Dangerous tools may run unsandboxed without user awareness. Mitigated by EventBus `SandboxDowngraded` events and `lango sandbox status` diagnostics.
- **[Profile temp file cleanup]** → Seatbelt profile written to /tmp, cleaned up via `CleanupProfileFile()` after process exit. Leaked files are small (~1KB) and in /tmp.
