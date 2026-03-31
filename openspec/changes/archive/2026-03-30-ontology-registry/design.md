## Context

Lango's graph layer stores triples in BoltDB with 9 hardcoded predicates (`store.go:11-21`) and untyped string node IDs using ad-hoc prefix conventions. The extractor (`extractor.go:114-120`) rejects unknown predicates via a compile-time switch. This architecture blocks domain-specific ontology growth — the agent cannot define new entity types or relationships.

This change introduces the ontology subsystem foundation: a Registry for ObjectType and PredicateDefinition metadata, and a ServiceImpl facade that all graph consumers will eventually use. In this change, the registry exists alongside the current graph code without modifying it.

**Current graph consumers** (will be migrated in Change 1-2):
- `graph.Extractor` — calls `isValidPredicate()` (hardcoded switch)
- `learning.GraphEngine` — creates triples with string prefix node IDs
- `memory.GraphHooks` — creates temporal triples
- `adk.ContextAwareModelAdapter` — queries graph RAG

## Goals / Non-Goals

**Goals:**
- Establish `internal/ontology/` package as the single authority for type/predicate metadata
- Persist ObjectType and PredicateDefinition in SQLite via Ent ORM
- Seed existing 9 predicates and 6 node types at startup (idempotent)
- Expose `PredicateValidator()` returning a cached `func(string) bool` for hot-path use
- Define `OntologyService` interface with extension points for future changes (1-3, 1-4, 1.5)
- Add `OntologyConfig` to config system

**Non-Goals:**
- Modifying existing graph store code (Change 1-2)
- Entity resolution / alias management (Change 1-4)
- Truth maintenance / temporal triples (Change 1-3)
- Structured property queries (Change 1.5-1)
- Ontology tools / ontologist agent (Change 1.5-2)

## Decisions

### D1: Facade pattern — OntologyService interface

**Decision**: All ontology consumers use a single `OntologyService` interface. Internal components (Registry, Resolver, TruthMaintainer, PropertyStore) are never referenced directly.

**Rationale**: BoltDB (relations) + SQLite (schema, properties, conflicts, aliases) creates storage complexity. A facade encapsulates this. Without it, graph/property/conflict/ACL logic leaks into app layer.

**Alternative considered**: Direct Registry access from consumers. Rejected because future changes add Resolver and TruthMaintainer — consumers would need to know about multiple subsystems.

### D2: Ent-backed Registry (SQLite)

**Decision**: Store ObjectType and PredicateDefinition in SQLite via Ent schemas, not in BoltDB.

**Rationale**: Schema metadata needs structured queries (list by status, lookup by name), versioning, and UNIQUE constraints. BoltDB's key-value model is poor for this. SQLite with Ent provides these natively.

**Alternative considered**: BoltDB bucket for schema. Rejected — querying by status/name requires full scan.

### D3: Cached PredicateValidator via closure

**Decision**: `PredicateValidator()` returns `func(string) bool` (context-free). Internally uses `map[string]bool` protected by `sync.RWMutex`. Cache refreshed on `RegisterPredicate`/`DeprecatePredicate`, not on TTL.

**Rationale**: `graph.Store.SetPredicateValidator` expects `func(string) bool` (hot path, called per triple). A DB query per validation is too expensive. Explicit invalidation is sufficient because predicate registration is rare (seed at startup, manual registration later).

**Alternative considered**: LRU cache with TTL. Rejected — unnecessary complexity for a set that changes rarely.

### D4: Idempotent seed migration

**Decision**: `SeedDefaults()` checks existence before inserting each type/predicate. Runs at every startup. No separate migration step.

**Rationale**: Keeps bootstrap simple. No ordering dependency with Ent auto-migration. Safe for repeated calls.

### D5: Cardinality includes ManyToOne

**Decision**: Four cardinality values: `OneToOne`, `OneToMany`, `ManyToOne`, `ManyToMany`. ManyToOne is semantically distinct from OneToMany (subject-perspective).

**Rationale**: `in_session` is ManyToOne (many entities → one session). Without ManyToOne, the seed would need to invert the predicate direction, which is confusing.

### D6: No graph store changes in this change

**Decision**: This change only creates the ontology package and seeds data. The graph store continues using hardcoded predicates. Integration happens in Change 1-2.

**Rationale**: Smaller rollback boundary. If the ontology registry has issues, existing graph functionality is unaffected.

## Risks / Trade-offs

- **[Risk] Ent code generation**: Adding 2 new schemas requires `go generate ./internal/ent`. If Ent version or generator config has issues, build breaks.
  → Mitigation: Run `go generate` immediately after schema creation, fix any issues before proceeding.

- **[Risk] Config field addition**: Adding `Ontology OntologyConfig` to `Config` struct may affect config serialization/deserialization.
  → Mitigation: Use `omitempty` JSON tag, default `Enabled: false`. Existing configs without the field deserialize cleanly.

- **[Risk] Seed data drift**: If graph predicates change in the future but seed isn't updated, registry becomes stale.
  → Mitigation: Seed references the same 9 const values from `graph.store.go`. Change 1-2 will replace the hardcoded validation with registry lookup.

- **[Trade-off] `StoreTriple` declared but trivially implemented**: In Change 1-1, `StoreTriple` just delegates to `store.AddTriple`. Full Resolve→Validate→Store pipeline comes in Change 1-4.
  → Acceptable: Interface stability is more important than implementation completeness at this stage.
