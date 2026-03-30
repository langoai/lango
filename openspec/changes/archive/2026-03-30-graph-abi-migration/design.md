## Context

Change 1-1 created `internal/ontology/` with `OntologyService`, `Registry`, seed migration (9 predicates + 6 types), and `PredicateValidator()` returning a cached `func(string) bool`. The graph layer is untouched — `graph.Triple` has no type fields, `Extractor` uses hardcoded `isValidPredicate`, and triple producers use string prefix conventions for node IDs.

This change bridges the ontology registry to the graph layer. After this change, every triple stored in BoltDB carries type metadata, and the Extractor validates predicates against the registry.

## Goals / Non-Goals

**Goals:**
- Add SubjectType/ObjectType to graph.Triple with full backward compatibility
- Wire predicate validation from ontology registry into BoltStore and Extractor
- Propagate type info through all producers (GraphEngine, MemoryHooks, wiring event handlers)
- Expose type info in GraphRAG results and graph tools

**Non-Goals:**
- Modifying the `graph.Store` interface (avoid mock/test ripple)
- Truth maintenance or temporal metadata (Change 1-3)
- Entity resolution or alias management (Change 1-4)
- Type-constraint validation on SourceTypes/TargetTypes (deferred — registry stores constraints but enforcement is Change 1-3+)

## Decisions

### D1: Store interface unchanged — validator on BoltStore concrete type only

**Decision**: `SetPredicateValidator(PredicateValidatorFunc)` is added to `BoltStore` struct, not the `Store` interface. Wiring uses concrete type assertion `gc.store.(*graph.BoltStore)`.

**Rationale**: `graph.Store` is implemented by `BoltStore`, `MockGraphStore` (`testutil/mock_graph.go`), and `fakeGraphStore` (`ontology_test.go`). Adding a method to the interface forces all three to update. Since validation is an optional enhancement (graceful fallback when nil), it belongs on the concrete type.

**Alternative**: Add to Store interface. Rejected — unnecessary ripple for optional behavior.

### D2: Type fields stored in metadata, not BoltDB key structure

**Decision**: SubjectType/ObjectType are stored as `_subject_type` / `_object_type` entries in the existing `Metadata map[string]string`. BoltDB key format (SPO/POS/OSP null-byte separated) is unchanged.

**Rationale**: Changing the key format would require data migration for existing BoltDB files. Metadata is already JSON-encoded as the value — adding two keys is zero-cost. Reading triples restores type fields from metadata automatically.

### D3: Extractor uses functional option pattern

**Decision**: `NewExtractor` gains variadic `...ExtractorOption`. `WithPredicateValidator(v)` injects the validator. Existing call sites that don't pass options continue to work with hardcoded fallback.

**Rationale**: The Extractor is constructed in `wiring_graph.go:76`. Adding a required parameter would be a breaking change. Functional options preserve backward compatibility and make the validator injection optional (ontology disabled → no option passed → hardcoded fallback).

### D4: Backward compatibility via zero-value fields

**Decision**: `SubjectType` and `ObjectType` default to empty string. All existing code that creates `graph.Triple{}` without these fields continues to work. Empty type fields are not stored in metadata (no `_subject_type: ""` entries).

**Rationale**: Go struct zero values. Existing triples in BoltDB have no `_subject_type` metadata key — on read, the type fields remain empty. No migration needed.

### D5: PredicateValidatorFunc defined in graph package

**Decision**: `type PredicateValidatorFunc func(name string) bool` is defined in `internal/graph/store.go`, not in the ontology package.

**Rationale**: The graph package should not import ontology (avoid circular dependency). The ontology package's `PredicateValidator()` returns `func(string) bool` which is assignable to `PredicateValidatorFunc`.

## Risks / Trade-offs

- **[Risk] Triple literal churn**: Adding two fields to Triple means every `graph.Triple{...}` literal in tests and production code gains two zero-value fields. Go allows omitting zero-value fields in struct literals with field names, so existing code compiles without changes.
  → Mitigation: Only update producers that should set type info. Test literals stay unchanged.

- **[Risk] Metadata key collision**: Using `_` prefix for system keys (`_subject_type`) could collide with user metadata.
  → Mitigation: Current code only uses `"source"` as metadata key. `_` prefix convention is established in this change and documented for future use.

- **[Trade-off] Validator in BoltStore but not in buffer path**: `GraphBuffer` enqueues triples and batch-writes via `store.AddTriples`. Validation happens in `putTriple` inside the BoltDB transaction. Invalid predicates are rejected at write time, not enqueue time.
  → Acceptable: Early rejection at enqueue would require the buffer to know about validation. Late rejection in putTriple is simpler and catches all paths.
