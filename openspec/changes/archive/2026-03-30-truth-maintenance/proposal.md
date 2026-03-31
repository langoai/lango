## Why

All triples in the graph store are "eternal" — once stored, they have no temporal dimension, no source attribution, and no confidence score. When contradictory facts appear (e.g., two different causes for the same error pattern on a OneToOne predicate), both remain valid with no mechanism to detect or resolve the conflict. This blocks Change 1-4 (Entity Resolution) whose Merge operation requires `RetractFact` to invalidate duplicate triples, and prevents any future time-travel queries or provenance tracking.

## What Changes

- Add bi-temporal metadata constants (`_valid_from`, `_valid_to`, `_recorded_at`, `_recorded_by`, `_source`, `_confidence`) to all new triples via the `TruthMaintainer` layer
- Implement `TruthMaintainer` interface with `AssertFact`, `RetractFact`, `ConflictSet`, `ResolveConflict`, `FactsAt`, `OpenConflicts`
- Add cardinality-based conflict detection: OneToOne predicates with different active objects trigger conflict creation
- Add source-of-truth auto-resolution via `sourcePrecedence` map (manual > knowledge > llm_extraction > graph_engine > memory_hook)
- Add `OntologyConflict` Ent schema for conflict persistence (SQLite)
- Extend `OntologyService` interface with 6 new truth maintenance methods
- Add reconciliation logic for cross-store consistency (SQLite conflict records vs BoltDB triples)

## Capabilities

### New Capabilities
- `truth-maintenance`: Bi-temporal metadata, conflict detection/resolution, fact assertion/retraction, time-travel queries

### Modified Capabilities
- `ontology-registry`: OntologyService interface extended with AssertFact, RetractFact, ConflictSet, ResolveConflict, FactsAt, OpenConflicts methods

## Impact

- `internal/ontology/` — new truth.go, truth_ent.go; service.go interface extension
- `internal/ent/schema/` — new ontology_conflict.go schema + Ent codegen
- `internal/graph/store.go` — no interface change; temporal metadata lives in existing `Metadata map[string]string`
- `internal/app/wiring_ontology.go` — TruthMaintainer initialization
- `internal/app/modules.go` — pass conflict store to ontology init
- Existing triple producers (GraphEngine, MemoryHooks, Extractor) are NOT modified in this change — they continue using `store.AddTriple` directly. Migration to `svc.AssertFact` is deferred to Change 1-4.
