## 1. Triple Struct Extension

- [x] 1.1 Add `SubjectType string` and `ObjectType string` fields to `graph.Triple` in `internal/graph/store.go`
- [x] 1.2 Add `PredicateValidatorFunc` type definition in `internal/graph/store.go`

## 2. BoltStore Type Metadata

- [x] 2.1 Add `validator PredicateValidatorFunc` field and `SetPredicateValidator` method to BoltStore in `internal/graph/bolt_store.go`
- [x] 2.2 Modify `putTriple` to: validate predicate (if validator set), inject `_subject_type`/`_object_type` into metadata before encoding
- [x] 2.3 Add `restoreTypeFields(*Triple)` helper and call it in `tripleFromSPOKey` and `tripleFromOSPKey` to restore SubjectType/ObjectType from metadata on read

## 3. Extractor Validator Injection

- [x] 3.1 Add `validator PredicateValidatorFunc` field to Extractor struct, convert `NewExtractor` to accept `...ExtractorOption`, add `WithPredicateValidator` option
- [x] 3.2 Convert standalone `isValidPredicate` to Extractor method with validator-or-fallback logic, rename original to `defaultIsValidPredicate`
- [x] 3.3 Update `parseResponse` to use `e.isValidPredicate()` (instance method) and log rejected predicates at warn level

## 4. Triple Producer Type Propagation

- [x] 4.1 Update `GraphEngine.recordErrorGraph`: set SubjectType/ObjectType on all created triples (ErrorPattern, Tool, Session)
- [x] 4.2 Update `GraphEngine.RecordFix`: set SubjectType/ObjectType on fix/session triples (Fix, Session)
- [x] 4.3 Update `MemoryGraphHooks.OnObservation`: set SubjectType/ObjectType (Observation, Session)
- [x] 4.4 Update `MemoryGraphHooks.OnReflection`: set SubjectType/ObjectType (Reflection, Session, Observation)
- [x] 4.5 Update `wiring_graph.go` ContentSavedEvent handler: set SubjectType/ObjectType on containment triples

## 5. EventBus Triple Mirror

- [x] 5.1 Add `SubjectType` and `ObjectType` fields to `eventbus.Triple` struct in `internal/eventbus/events.go`
- [x] 5.2 Update `wiring_graph.go` TriplesExtractedEvent subscriber: copy SubjectType/ObjectType in graph↔eventbus Triple conversion

## 6. GraphRAG Type Propagation

- [x] 6.1 Add `NodeType string` field to `GraphNode` in `internal/graph/rag.go`
- [x] 6.2 Populate NodeType from triple metadata during graph expansion traversal
- [x] 6.3 Update `AssembleSection` to format typed nodes as `**NodeType:ID**` when NodeType is non-empty

## 7. Graph Tools Type Filter

- [x] 7.1 Add optional `node_type_filter` parameter to `graph_traverse` tool, post-filter results by SubjectType
- [x] 7.2 Add optional `subject_type` and `object_type` parameters to `graph_query` tool, post-filter results

## 8. Wiring Integration

- [x] 8.1 Update `wireGraphCallbacks` signature to accept `ontology.OntologyService`, extract `PredicateValidator()` closure
- [x] 8.2 Inject validator into BoltStore via concrete type assertion `gc.store.(*graph.BoltStore)`
- [x] 8.3 Pass validator to Extractor via `graph.WithPredicateValidator(validator)` option
- [x] 8.4 Update `modules.go` intelligence module to pass `ontologySvc` to `wireGraphCallbacks`

## 9. Tests and Verification

- [x] 9.1 Unit test: BoltStore round-trip with typed triple (store → query → verify SubjectType/ObjectType)
- [x] 9.2 Unit test: BoltStore predicate validation (valid accepted, invalid rejected, no validator = all accepted)
- [x] 9.3 Unit test: BoltStore backward compat (untyped triple stores/retrieves without error)
- [x] 9.4 Unit test: Extractor with validator rejects unknown predicate, without validator uses fallback
- [x] 9.5 Verify `go build -tags fts5 ./...` succeeds
- [x] 9.6 Verify `go test -tags fts5 ./internal/graph/... -v` passes (no regression)
- [x] 9.7 Verify `go test -tags fts5 ./internal/learning/... -v` passes
- [x] 9.8 Verify `go test -tags fts5 ./internal/memory/... -v` passes
- [x] 9.9 Verify `go test -tags fts5 ./internal/ontology/... -v` passes (no regression)
