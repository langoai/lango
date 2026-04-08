## Design Notes

### D1: bwrap probe is two-phase (base + network)

**Problem**: `NewBwrapIsolator()` only ran `exec.Command(abs, "--version").Output()`, which reports the binary metadata regardless of namespace-creation permission. On hardened distros (`kernel.unprivileged_userns_clone=0`, AppArmor lockdown, missing setuid-root) this returns exit 0 but the real `Apply()` → `cmd.Run()` later fails with EPERM — fail-closed mode then rejects every exec/skill/MCP command.

**Decision**: Split the probe into two phases that both reuse `compileBwrapArgs` for argv generation, so probe and runtime cannot drift:

1. **Base probe** runs `compileBwrapArgs(Policy{Filesystem:{ReadOnlyGlobal:true}, Network:NetworkAllow, Process:{AllowFork:true}})` + `-- /bin/true`. `NetworkAllow` matches `MCPServerPolicy` exactly — the absolute minimum every lango consumer needs. Failure marks the isolator fully unavailable.
2. **Network probe** runs the same shape but with `Network:NetworkDeny`, which `compileBwrapArgs` translates to `--unshare-net`. Failure does NOT change `Available()`; instead, a separate `NetworkIsolationAvailable() bool` + `NetworkIsolationReason() string` pair is exposed, and `Apply()` rejects `NetworkDeny`/`NetworkUnixOnly` policies at runtime with an `ErrIsolatorUnavailable`-wrapped error. `NetworkAllow` policies (MCP) are unaffected.

**Why two phases and not one probe with `NetworkDeny`**: `MCPServerPolicy` uses `NetworkAllow`, so hosts that permit pid/ipc/uts/root-bind but block `--unshare-net` (e.g. Docker with `--cap-drop=NET_ADMIN`, some CI runners) can run MCP fine. A single probe with `NetworkDeny` would over-reject — the whole isolator would be marked unavailable even though MCP would work. The two-phase design keeps each consumer's failure mode precise.

**Why reuse `compileBwrapArgs` instead of hand-listing flags**: Hand-listed probe flags drift from runtime as `compileBwrapArgs` evolves. The memory feedback `probe-matches-runtime-flags` (written during this plan's review rounds) records two concrete drift incidents from the draft:
- Round 1 draft included `--unshare-user`, which `compileBwrapArgs` does NOT use; probing with `--unshare-user` would false-negative setuid-root bwrap on hardened distros.
- Round 2 draft used `NetworkDeny` for the single-phase probe, which false-negated MCP consumers.

Reusing `compileBwrapArgs` closes both drift vectors structurally — probe and runtime share one argv generator.

**Why `Reason()` stays empty when `Available()==true`**: The existing `linux-bwrap-isolation` spec requires `Reason() SHALL return ""` when the isolator is available. Partial degradation is surfaced through the dedicated `NetworkIsolationReason()` method so the existing contract remains intact and older consumers continue to work unchanged.

**Probe timeout**: 2 seconds per phase, 4 seconds max total. Base failure short-circuits (no network probe). Normal-environment measurement: < 100ms total for both phases.

### D2: `.git` walk-up discovery via private helper

**Problem**: `DefaultToolPolicy` used `filepath.Join(workDir, ".git")` at the immediate level. Supervisor/skill executor pass the supervisor cwd, which is commonly a subdirectory of the actual repo (`/repo/cmd/lango` while `.git` lives at `/repo/.git`). The `isDir` guard silently dropped the deny, leaving sandboxed children free to read git metadata.

**Decision**: Add a private `findGitRoot(workDir)` helper in `internal/sandbox/os/policy.go` that walks upward from `workDir` to the first ancestor containing a `.git` directory. Termination: `filepath.Dir(cur) == cur` (filesystem root). Returns "" if no `.git` ancestor found — callers simply drop the baseline deny.

**Worktree trade-off**: A linked worktree stores `.git` as a regular file (`gitdir: <path>`). `findGitRoot` skips file entries because `compileBwrapArgs` cannot mount `--tmpfs` on a regular file, and walk-up continues past it. Alternatives considered:
- *Parse the gitdir pointer*: adds filesystem-reading code path inside policy construction, increases attack surface for symlink-chase bugs, and doesn't solve the underlying "deny a file" requirement.
- *File-level deny via `--ro-bind /dev/null <file>`*: this is the correct long-term fix and lands in PR 5c. PR 5a holds the line with a documented gap.

Worktree users currently retain the existing `dataRoot` (control-plane) protection and the workspace `WritePaths` boundary, so the lango-specific leak goals (control-plane, session DB, audit log) are still met. Only the repo's own git metadata leaks to sandboxed children in worktree mode.

**Why not `filepath.EvalSymlinks`**: PR 5c territory. Adding symlink resolution to `findGitRoot` alone would be half a solution — `compileBwrapArgs` and Seatbelt profile generation also need symlink-aware normalization. Keep the boundary clean by leaving all symlink work to PR 5c.

### D3: `MCPServerPolicy(workDir, dataRoot)` signature extension

**Problem**: `MCPServerPolicy(dataRoot string)` did not know about the workspace. MCP stdio children had no `.git` baseline deny — asymmetric with `DefaultToolPolicy`. The memory feedback `sandbox-apply-has-three-call-sites` explicitly warned that cross-cutting policy changes must update exec/skill/MCP simultaneously; PR 4 round 2 fixed `DefaultToolPolicy` but left MCP behind.

**Decision**: Extend the signature to `MCPServerPolicy(workDir, dataRoot string)` and call `findGitRoot(workDir)` for the baseline deny. Symmetric with the updated `DefaultToolPolicy`. Wire `workspacePath` through `ServerManager.SetOSIsolator` → `ServerConnection.SetOSIsolator` → `createTransport`. In `createTransport`, also set `cmd.Dir = sc.workspacePath` (when non-empty) so the MCP child runs with cwd inside the user's workspace — policy discovery and execution share the same git context.

**Legacy fallback**: Empty `workspacePath` leaves `cmd.Dir` unset (Go's `exec.Cmd` default behavior: inherit supervisor cwd). Existing callers that don't migrate stay functional. New app wiring in `internal/app/wiring_mcp.go` resolves `cfg.Sandbox.WorkspacePath` with `os.Getwd()` fallback — identical pattern to supervisor and skill registry, so the three Apply() sites now share the same workspacePath resolution idiom.

### D4: CLI status surfaces partial degradation

**Problem**: When `NetworkIsolationAvailable()==false` but base availability is true, a user whose MCP servers work while exec/skill commands get rejected has no diagnostic trail. The existing `Active Isolation` section shows `Available: true` and stops.

**Decision**: Define a local `networkIsolator` interface in `internal/cli/sandbox/sandbox.go` (mirroring the existing `versioner` pattern) and surface a single `Network Iso: unavailable (reason)` line when the isolator implements the interface AND reports the partial degradation. Clean-state isolators omit the line entirely — no noise when everything works.

Interface lives in the CLI package because no other consumer needs to type-assert to it; keeps the `OSIsolator` core interface minimal.

## Worktree Trade-off (Explicit Note)

Linked git worktrees use a regular file (`.git: gitdir: <path>`) as the marker instead of a directory. `findGitRoot` skips these because bwrap's `--tmpfs` cannot mount on a file. Walk-up continues past the worktree pointer; if the walk still terminates at filesystem root with no ancestor `.git` directory, no baseline deny is added.

This is a regression from PR 4 only in the sense that a worktree previously with an immediate `.git` file would also have been skipped (the old code had the same gap via a different mechanism). PR 5c closes the file-level deny gap and restores protection to worktrees.

## Alternatives Considered

- **Defer the whole probe to `Apply()` time, cache per-policy**: changes the availability semantics, breaks the existing `Available() bool` contract, and would require every consumer to re-interpret what "available" means. Rejected — two-phase probe at startup is simpler and preserves compatibility.
- **Add a `backend=auto` fallback that selects seatbelt → bwrap → native → noop**: orthogonal to this change. The current `SelectBackend` already does this.
- **Silently use old behavior when the probe fails**: hides the problem and keeps users in the confusing "sandbox OK but commands denied" state. Rejected.
- **Bundle PR 5a with the OpenSpec spec format meta-fix**: would violate `openspec-change-boundaries-v2` — they have no causal relationship. Kept separate as PR 5b.
