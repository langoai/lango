## Context

Changes 2-1 (ACL) and 2-2 (Actions) are complete. OntologyService now has 33 methods, ACL guards, ActionExecutor, ActionLogStore. ServiceImpl has 11 fields. SchemaStatus has `active` and `deprecated`. Ent schemas for `ontology_type` and `ontology_predicate` use string-based enum with those 2 values.

## Goals / Non-Goals

**Goals:**
- 5-state lifecycle FSM for types and predicates (proposed → shadow → active → deprecated, with quarantined branch)
- Rate limiting to prevent schema explosion
- Schema health monitoring
- SeedDefaults bypass via wiring order
- Backward compatible: governance nil = existing behavior

**Non-Goals:**
- Automatic time-based promotion (future enhancement)
- Usage tracking/counting (TypeUsage returns stub for now — full implementation requires event tracking)
- Per-type or per-predicate governance policies (single global policy)

## Decisions

### D1: GovernanceEngine is kind-agnostic

FSM rules are identical for types and predicates. `ValidateTransition(from, to)` doesn't need a `kind` parameter. Rate limiting is a combined daily count (type + predicate together).

**Why:** Schema explosion is kind-agnostic — 50 new types or 50 new predicates are equally problematic. Separate limits add complexity without benefit.

### D2: RegisterType/RegisterPredicate force `proposed` when governance enabled

When `s.governance != nil`, the status is overridden to `SchemaProposed` regardless of input. Rate limit is checked before registration.

**Why:** Governance must be opt-in but non-bypassable. When enabled, no path should create `active` types without going through the FSM.

### D3: SeedDefaults bypass via wiring order

`SeedDefaults` runs before `SetGovernanceEngine` in `wiring_ontology.go`. This means seed types/predicates register with `active` status normally, then governance is enabled for all subsequent registrations.

**Why:** Same pattern as truth maintainer, entity resolver, ACL — setter injection after construction. No special bypass logic needed.

### D4: Shadow predicates included in predicate cache

`refreshPredicateCache` includes both `SchemaActive` and `SchemaShadow` predicates. This means shadow predicates pass validation and can be used in triples.

**Why:** Shadow = "trial operation" — the predicate is being tested with real data before full promotion. Proposed and quarantined predicates should NOT be usable.

### D5: TypeUsage as stub

`TypeUsage` returns basic info (status, created date) but not actual usage counts. Full usage tracking would require counting triples/properties per type, which is expensive and belongs in a future observability change.

**Why:** The interface is defined now for tool generation, but implementation is deliberately minimal.

## Risks / Trade-offs

| Risk | Mitigation |
|------|-----------|
| Ent enum expansion breaks existing data | SQLite uses string columns — additive enum is safe, no migration needed |
| Governance changes RegisterType behavior | governance nil = existing behavior, tests don't set governance |
| Shadow predicates leak unvalidated schemas | Shadow is intentional "trial" — only proposed/quarantined are blocked |
| Rate limit resets on restart (in-memory) | Acceptable for v1 single-node; Ent-backed tracking for future |
