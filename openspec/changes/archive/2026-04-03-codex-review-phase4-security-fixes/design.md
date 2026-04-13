## Context

Codex automated review of the Phase 4 branch diff against `dev` found 3 security issues across 4 review rounds. Each fix was iteratively refined based on subsequent review feedback.

## Goals / Non-Goals

**Goals:**
- Fix DNS rebinding bypass in P2P browser navigation
- Restore correct `toolIsolation.enabled` semantics for P2P sandbox executor
- Make filesystem `Delete` safe for symlinks: delete the link, not the target
- Handle edge cases: OS path aliases (macOS `/var` → `/private/var`) and symlink-as-config-entry

**Non-Goals:**
- Adding new filesystem operations or browser features
- Changing the P2P tool isolation default (remains `false` — safe posture)

## Decisions

### 1. Always re-validate post-navigation URL in P2P context

**Decision:** Remove `finalURL != rawURL` condition. Always call `ValidateURLForP2P(finalURL)` after navigation.

**Rationale:** DNS rebinding attacks produce the same URL string but different IP resolution at navigation time. String comparison cannot detect this.

### 2. Restore toolIsolation.enabled gate with startup warning

**Decision:** Keep `if cfg.P2P.ToolIsolation.Enabled` gate. Add `else` warning log.

**Rationale:** First review said removing the gate breaks P2P. Second review said removing it silently enables remote execution. The handler's `sandboxExec == nil → reject` is the intentional safe default. The config switch must be respected. Warning log helps operators understand why inbound tool calls are rejected.

### 3. Symlink-specific Delete flow (Lstat before resolve)

**Decision:** Check `os.Lstat` before calling `validatePath`. If symlink: resolve only parent directory with `EvalSymlinks`, validate canonical link location, delete the link. If not symlink: use standard `validatePath` flow.

**Rationale:** Delete is fundamentally different from Read/Write — removing a symlink doesn't touch the target, so the target's security classification is irrelevant. Only the link's location matters. Parent-only resolution handles OS aliases without resolving the link itself.

### 4. Dual comparison in checkPathAccess

**Decision:** Compare input path against both unresolved and resolved versions of each config entry.

**Rationale:** If an operator configures `BlockedPaths: ["/tmp/secret-link"]` where the entry itself is a symlink, only resolving the config entry (to the target) would miss the string match. Checking both covers all cases.

## Risks / Trade-offs

- **[Symlink to blocked target cleanup]** Intentionally allowed: `workspace/link → /etc/passwd` can now be deleted (removes the link, not `/etc/passwd`). This is the desired behavior — operators can clean up dangerous symlinks.
- **[P2P tool calls rejected by default]** Intentional: `toolIsolation.enabled=false` means no remote execution. Startup warning makes this visible.
