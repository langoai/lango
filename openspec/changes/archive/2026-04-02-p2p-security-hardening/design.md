## Context

P2P tool execution has 5 fail-open paths where remote peers can bypass intended security boundaries. The sandbox, filesystem, browser, and protocol handler layers each have independent gaps that combine to create a significant attack surface for any P2P-enabled deployment.

## Goals / Non-Goals

**Goals:**
- All P2P tool invocations pass through sandbox, safety gate, and context-aware tool restrictions
- Default config is fail-closed for new installations (existing deployments can opt out)
- No import cycles between `p2p/protocol`, `agent`, and `toolcatalog`
- Backward compatible — nil checker / disabled gates pass all tools through

**Non-Goals:**
- Changing the sandbox runtime probe chain (Docker → gVisor → Native)
- Adding gVisor implementation (remains a stub)
- Modifying firewall ACL logic (orthogonal to safety gate)
- Per-tool P2P policies beyond SafetyLevel (future work)

## Decisions

1. **SafetyLevelChecker as callback type** — Use `func(toolName string) (int, bool)` in the handler package instead of importing `agent.SafetyLevel` directly. Rationale: avoids import cycle `p2p/protocol → agent → ...`.

2. **Numeric safety levels** — Map Safe=1, Moderate=2, Dangerous=3 for comparison. `ParseSafetyLevel` maps unknown strings to Dangerous (fail-safe).

3. **Separate P2P context keys** — `filesystem.WithP2PContext`/`IsP2PContext` (package-level) for filesystem-specific P2P detection, `ctxkeys.WithP2PRequest`/`IsP2PRequest` for cross-package propagation. Rationale: filesystem context is self-contained; browser/handler need cross-package key.

4. **EvalSymlinks before allowed-path check** — Resolve symlinks for both the target path AND the allowed/blocked config paths. Rationale: config paths like `/var` on macOS resolve to `/private/var`.

5. **Single-file delete in P2P** — `os.Remove` instead of `os.RemoveAll` when `IsP2PContext(ctx)`. Rationale: prevents recursive deletion of directory trees by remote peers while still allowing legitimate file cleanup.

6. **Private network CIDR blocklist** — Hardcoded `privateNetworks` slice with defense-in-depth `ip.IsLoopback()` check. Rationale: standard private ranges are well-defined; no config needed.

## Risks / Trade-offs

- [RequireContainer default=true may break existing P2P deployments without Docker] → Mitigated by config override `p2p.toolIsolation.requireContainer: false`
- [SafetyLevel gate only checks numeric level, not tool-specific context] → Acceptable for v1; firewall ACL provides tool-name-level control
- [EvalSymlinks on broken symlinks returns error, could reject valid write targets] → Mitigated by continuing with cleaned absolute path when EvalSymlinks fails
- [URL validator only checks IP literals, not DNS-resolved addresses] → Acceptable trade-off; DNS resolution in the validator path would add latency and complexity

## Data Flow

```
P2P Request → Handler
  ├─ ctxkeys.WithP2PRequest(ctx)
  ├─ checkSafetyGate(toolName) → deny if level > maxSafetyLevel
  ├─ sandboxExec == nil? → deny with ErrNoSandboxExecutor
  ├─ Firewall ACL check
  ├─ Owner approval check
  └─ Execute tool with P2P context
       ├─ browser: ValidateURLForP2P + block eval
       ├─ filesystem: EvalSymlinks + single-file delete
       └─ sandbox: RequireContainer fail-closed
```
