## Why

Three deferred gaps from PR 4 (`sandbox-escape-hardening`) still violate the contract that PR promised. This change closes them as a single causally-linked fix PR (PR 5a in the sandbox roadmap), leaving PR 5b (OpenSpec spec format cleanup), PR 5c (file-level deny + symlink + glob), and PR 5d (native Linux backend) for later.

1. **`bwrap --version` false-positive availability**: `NewBwrapIsolator` probes only the binary metadata, so hosts where `bwrap --version` succeeds but kernel namespace creation is blocked (e.g. Debian/Ubuntu with `kernel.unprivileged_userns_clone=0`, AppArmor lockdown, or missing setuid-root) are advertised as `Available()==true`. Fail-closed mode then rejects every exec/skill/MCP command at first invocation, confusing users who see "sandbox is OK but all commands are denied".
2. **`.git` walk-up not performed**: `DefaultToolPolicy` only checks `filepath.Join(workDir, ".git")` at the immediate level. When supervisor or skill executor passes a subdirectory as `workDir` (e.g. cwd = `/repo/cmd/lango` while `.git` lives at `/repo/.git`), the `isDir` guard silently drops the deny and sandboxed children can read git metadata freely.
3. **`MCPServerPolicy` workspace asymmetry**: `MCPServerPolicy(dataRoot)` takes no workspace argument, so MCP stdio children have no `.git` baseline deny — asymmetric with `DefaultToolPolicy`. The memory feedback `sandbox-apply-has-three-call-sites` warned that cross-cutting policy changes must update exec/skill/mcp simultaneously; PR 4 missed the MCP side.

## What Changes

- **BREAKING (internal API)**: `MCPServerPolicy(dataRoot string)` → `MCPServerPolicy(workDir, dataRoot string)`. Internal only — no external/API consumers. All three policy helpers (`DefaultToolPolicy`, `StrictToolPolicy`, `MCPServerPolicy`) now share a symmetric baseline deny shape.
- **BREAKING (internal API)**: `ServerManager.SetOSIsolator(iso, dataRoot)` → `SetOSIsolator(iso, workspacePath, dataRoot)` and `ServerConnection.SetOSIsolator(iso, dataRoot)` → `SetOSIsolator(iso, workspacePath, dataRoot)`. App wiring passes `cfg.Sandbox.WorkspacePath` with `os.Getwd()` fallback — identical pattern to supervisor and skill registry.
- `NewBwrapIsolator` now runs a two-phase smoke probe: base probe (NetworkAllow, matches `MCPServerPolicy`) validates kernel namespace creation; network probe (NetworkDeny, matches `DefaultToolPolicy`) additionally validates `--unshare-net`. Both probes reuse `compileBwrapArgs` so probe argv cannot drift from the runtime `Apply()` path.
- Base probe failure marks the whole isolator unavailable (matches existing contract). Network probe failure only downgrades `Apply()` to reject `NetworkDeny`/`NetworkUnixOnly` policies with an `ErrIsolatorUnavailable`-wrapped error; `NetworkAllow` policies (MCP) continue to work.
- New `BwrapIsolator.NetworkIsolationAvailable() bool` and `NetworkIsolationReason() string` methods expose partial degradation without breaking the existing `Reason()` contract (which remains empty when `Available()==true`).
- `DefaultToolPolicy` now uses a private `findGitRoot(workDir)` helper that walks up from `workDir` to the first ancestor `.git` directory. Worktree pointers (`.git` as a regular file) are skipped because `compileBwrapArgs` cannot mount `--tmpfs` on a file; that gap closes in PR 5c with file-level deny semantics.
- `MCPServerPolicy` now also calls `findGitRoot(workDir)` so MCP stdio children get the same baseline deny as skills and exec tools.
- `createTransport` in `internal/mcp/connection.go` sets `cmd.Dir = sc.workspacePath` when non-empty so policy discovery and execution share the same git context. Empty `workspacePath` falls back to supervisor cwd (legacy behavior).
- `lango sandbox status` Active Isolation section gains a `Network Iso: unavailable (reason)` line when the bwrap network probe fails — surfaces partial degradation so users can diagnose "MCP works but exec/skill rejected" flows.
- `docs/cli/sandbox.md` gains one-line mentions of the smoke probe and the network isolation partial degradation UX.

## Capabilities

### New Capabilities
None.

### Modified Capabilities
- `linux-bwrap-isolation`: bwrap availability contract now requires a two-phase namespace smoke probe (base + network), not just `bwrap --version`. Adds partial-degradation semantics via `NetworkIsolationAvailable`/`NetworkIsolationReason`.
- `os-sandbox-core`: `DefaultToolPolicy` walks up to find the ancestor `.git` directory. `MCPServerPolicy` gains a `workDir` parameter and applies the same walk-up `.git` baseline deny as `DefaultToolPolicy`.
- `os-sandbox-integration`: `ServerConnection` carries a `workspacePath` field; `SetOSIsolator` takes `(iso, workspacePath, dataRoot)`; `createTransport` sets `cmd.Dir = workspacePath` when non-empty and passes it to `MCPServerPolicy`.

Note: `mcp-integration/spec.md` has a redundant one-line reference to `MCPServerPolicy(dataRoot)` inside a sub-section outside its `## Requirements` header. The openspec archive matcher cannot locate requirements under sub-sections, so that stale reference is intentionally NOT updated in this change. PR 5b (OpenSpec spec format meta-fix) will restructure `mcp-integration/spec.md` and close the drift.

## Impact

**Code**:
- `internal/sandbox/os/bwrap_linux.go` (two-phase probe + Apply gate + new methods)
- `internal/sandbox/os/bwrap_linux_test.go` (probe contract + Apply gate tests)
- `internal/sandbox/os/policy.go` (findGitRoot helper + DefaultToolPolicy walk-up + MCPServerPolicy signature)
- `internal/sandbox/os/policy_test.go` (findGitRoot table + walk-up regression + MCPServerPolicy 2-arg tests)
- `internal/sandbox/os/bwrap_args_test.go` (MCPServerPolicy 2-arg update)
- `internal/mcp/connection.go` (workspacePath field + setter + createTransport cmd.Dir + 2-arg policy call)
- `internal/mcp/manager.go` (workspacePath field + propagation)
- `internal/mcp/connection_test.go` (3-arg SetOSIsolator calls)
- `internal/app/wiring_mcp.go` (workspacePath resolution with os.Getwd fallback)
- `internal/cli/sandbox/sandbox.go` (networkIsolator interface + status line)

**Docs**: `docs/cli/sandbox.md` (two one-line additions).

**Behavior**: Hardened distros that previously false-positive on bwrap are now correctly marked unavailable with an actionable reason. Subdirectory-cwd workspaces gain `.git` protection for the first time. MCP stdio children are now protected from reading workspace git metadata. Partial bwrap degradation (network-only failure) preserves MCP usability while rejecting exec/skill NetworkDeny policies.

**Dependencies**: None new.
