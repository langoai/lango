## Context

Change 2-1 (ontology-acl) is complete. OntologyService now has 29 methods with ACL guards, `ACLPolicy` interface, `RoleBasedPolicy`, and principal context propagation via `mw_principal.go`. ServiceImpl has 9 fields (added `acl`).

This change adds a transactional action layer on top of OntologyService. Actions are in-process Go closures — not persisted DSL. The registry is populated at startup.

## Goals / Non-Goals

**Goals:**
- Reusable, composable transaction units with Precondition → Execute → Compensate lifecycle
- ACL enforcement at executor level (direct ACLPolicy injection) with defense-in-depth at service layer
- Structured execution logging via Ent-backed ActionLogStore
- Dynamic tool generation from ActionRegistry for agent consumption
- 2 built-in seed actions demonstrating the pattern

**Non-Goals:**
- Persisted/serializable action definitions (future Stage 3 for P2P exchange)
- Distributed transactions or saga orchestration across services
- Action versioning or migration
- Undo/redo beyond single-action compensation

## Decisions

### D1: ActionType is in-process built-in registry, NOT persisted DSL

ActionType uses Go closure fields (Precondition, Execute, Compensate). Cannot be serialized or stored externally. Registry is populated programmatically at startup.

**Why:** Closure-based actions are type-safe, testable, and don't require a DSL parser. P2P action exchange (Stage 3) will need a separate declarative ActionSpec if/when required.

### D2: ActionExecutor injects ACLPolicy directly (not via OntologyService)

ActionExecutor has its own `acl ACLPolicy` field. Does NOT call `OntologyService.CheckPermission` — that method doesn't exist on the interface.

**Why:** ACLPolicy is an independent interface. Direct injection follows the existing setter pattern and avoids adding internal concerns to the public interface. Same ACLPolicy instance is shared between ServiceImpl and ActionExecutor via wiring.

### D3: RequiredPerm invariant — executor perm >= inner service method perm

`ActionType.RequiredPerm` MUST be >= the maximum permission required by any OntologyService method called within Execute or Compensate. This prevents "executor passed but service rejected" partial failures.

**Why:** ActionExecutor checks RequiredPerm before invoking Execute. OntologyService methods perform their own ACL checks (defense in depth). If RequiredPerm < inner method perm, the executor passes but inner rejects, causing confusing partial failures.

### D4: BuildTools signature change — reg parameter added

`BuildTools(svc OntologyService, reg *ActionRegistry)`. When reg is nil, only 13 static tools are returned (backward compat for tests). When non-nil, dynamic `ontology_action_{name}` tools are appended.

**Why:** Dynamic tools need access to the registry. nil-safe default preserves existing test code with minimal changes.

### D5: ActionLog is separate from AuditLog

ActionLog has status FSM (started → completed/failed/compensated), structured params/effects/error fields. AuditLog has fixed action enum + generic details JSON. Different lifecycle requirements justify separate Ent schema.

**Why:** AuditLog is append-only events. ActionLog tracks mutable execution state with status transitions and compensation records.

## Risks / Trade-offs

| Risk | Mitigation |
|------|-----------|
| RequiredPerm invariant violation in future actions | Code comment on RequiredPerm field, test verifying built-in actions |
| BuildTools signature change breaks callers | Only 2 call sites: modules.go, tools_test.go. reg=nil backward compat |
| ActionType closure fields not JSON-serializable | By design — ActionLog stores params/effects as JSON, not the closures |
| Ent codegen required after schema addition | Documented in tasks, run before build |
