## Context

OntologyService (29 methods, `internal/ontology/service.go:17-69`) is the single facade for all ontology operations. Currently any caller — agent tools, programmatic code, SeedDefaults — has unrestricted access to all operations including destructive ones (MergeEntities, SplitEntity, DeleteEntityProperties).

The multi-agent orchestrator (`internal/orchestration/`) assigns tools by agent name: "ontologist" gets `ontology_*` tools, "chronicler" gets `memory_*` tools. Agent name is injected into context at `internal/adk/tools.go:172` via `ctxkeys.WithAgentName()`. This existing context key is the natural source for principal identity.

## Goals / Non-Goals

**Goals:**
- Fine-grained permission control (read/write/admin) on all OntologyService methods
- Principal derived from agent name via middleware, with "system" default for programmatic paths
- Backward compatible: ACL disabled by default (nil policy = allow all)
- Configurable role mapping via `lango.json`

**Non-Goals:**
- Per-entity or per-type access control (row-level security)
- User-level authentication (this is agent/principal-level)
- Tool-level blocking (that's `toolchain/hook_access.go` — coarse, agent-level)
- Dynamic role changes at runtime

## Decisions

### D1: ACL at service layer, not tool hook

ACL checks inside `ServiceImpl.checkPermission()`, not in toolchain middleware.

**Why:** Programmatic callers (SeedDefaults, ActionExecutor in 2-2, internal wiring) bypass tool middleware entirely. Service-layer ACL catches all call paths. Tool hooks (`hook_access.go`) control which agents can invoke which tools — a different concern.

**Alternative:** Toolchain middleware with permission annotations per tool. Rejected: doesn't cover programmatic callers, couples ACL to tool registration.

### D2: Separate `lango.principal` context key with middleware bridge

Add `WithPrincipal`/`PrincipalFromContext` in `ctxkeys` (separate from `agentNameKey`). A new `mw_principal.go` middleware copies agent name → principal at B4c2 position.

**Why:** Principal is semantically distinct from agent name. Future divergence (user-level principals, P2P peer identity) won't require rewiring. The middleware makes the injection point explicit and auditable.

**Alternative:** Reuse `AgentNameFromContext()` directly in `checkPermission`. Rejected: implicit coupling, no explicit injection point, harder to extend.

### D3: Permission as int with ordered comparison

`type Permission int` with `PermRead=1 < PermWrite=2 < PermAdmin=3`. Policy check: `roles[principal] >= required`.

**Why:** Simple ordered comparison replaces set-membership checks. Config stores strings ("read"/"write"/"admin"), converted at wiring time.

**Alternative:** String-based permission with explicit hierarchy map. Rejected: more code for same semantics.

### D4: Unknown principal defaults to PermRead, empty/"system" defaults to PermAdmin

- `""` or `"system"` → PermAdmin (programmatic paths: SeedDefaults, internal wiring, tests)
- Unknown principal (not in roles map) → PermRead (safe default for unexpected callers)

**Why:** "system" principal only occurs on programmatic paths without agent context. Tool execution always passes through WithPrincipal middleware, so "" never reaches service layer from user-facing paths. Unknown principals get read-only as defense-in-depth.

### D5: PredicateValidator excluded from ACL

`PredicateValidator() func(name string) bool` has no `ctx` parameter — returns a closure for hot-path validation. Cannot enforce ACL.

**Why:** This is a perf optimization returning a cached closure. Changing its signature to accept ctx would break the hot-path contract. Read-only semantics make this acceptable.

## Risks / Trade-offs

| Risk | Mitigation |
|------|-----------|
| ACL guard omitted on new methods added in 2-2/2-3 | Test that verifies all interface methods (except PredicateValidator) have guards |
| Middleware position change breaks principal flow | Integration test: tool call → service ACL check → principal matches agent name |
| "system" default too permissive | Only reachable from programmatic paths; tool middleware always sets principal |
| Permission model too coarse (3 levels) | Sufficient for Stage 2; per-operation policies can layer on top later |
