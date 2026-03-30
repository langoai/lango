## Why

Change 1-1 established the ontology registry (`internal/ontology/`) with typed ObjectTypes and PredicateDefinitions, but the graph layer still uses untyped `graph.Triple` with string-only Subject/Predicate/Object fields (`store.go:38-43`). Triple producers (Extractor, GraphEngine, MemoryHooks) create nodes with ad-hoc string prefix conventions (`error:`, `tool:`, `observation:`) and the Extractor validates predicates against a hardcoded switch (`extractor.go:114-120`). This change connects the graph layer to the ontology registry by adding type fields to Triple and wiring predicate validation through the registry.

## What Changes

- Add `SubjectType` and `ObjectType` string fields to `graph.Triple` (backward compatible — empty = untyped)
- Store type info in BoltDB triple metadata (`_subject_type`, `_object_type`) and restore on read
- Add `PredicateValidatorFunc` type and `SetPredicateValidator` method to `BoltStore` (concrete type only, **not** the `Store` interface)
- Replace `extractor.go`'s hardcoded `isValidPredicate` with ontology-backed validator via functional option `WithPredicateValidator`
- Add `SubjectType`/`ObjectType` to all triple producers: `GraphEngine.recordErrorGraph/RecordFix`, `MemoryGraphHooks.OnObservation/OnReflection`, `wiring_graph.go` content.saved handler
- Add `SubjectType`/`ObjectType` to `eventbus.Triple` mirror and update conversion code
- Add `NodeType` field to `graph.GraphNode` and populate from metadata during RAG traversal
- Add optional `node_type_filter` parameter to `graph_traverse` and `graph_query` tools
- Wire ontology service validator into BoltStore and Extractor in `wiring_graph.go`/`modules.go`

## Capabilities

### New Capabilities
- `graph-abi-typed-triple`: Typed Triple struct with SubjectType/ObjectType fields, BoltStore metadata persistence, registry-backed predicate validation, and type propagation across all graph producers

### Modified Capabilities
- `graph-store`: Triple struct gains SubjectType/ObjectType fields, BoltStore gains SetPredicateValidator
- `graph-rag`: GraphNode gains NodeType field populated from triple metadata

## Impact

- Modified: `internal/graph/store.go` (Triple struct), `bolt_store.go` (storage/retrieval + validator), `extractor.go` (validator injection), `rag.go` (GraphNode + type extraction), `tools.go` (type filter param)
- Modified: `internal/learning/graph_engine.go` (type fields on all created triples)
- Modified: `internal/memory/graph_hooks.go` (type fields on all created triples)
- Modified: `internal/eventbus/events.go` (Triple mirror + type fields)
- Modified: `internal/app/wiring_graph.go` (ontology service injection), `modules.go` (pass ontologySvc)
- **Not modified**: `internal/graph/store.go` Store interface, `internal/testutil/mock_graph.go`, `internal/ontology/ontology_test.go` (no interface ripple)
