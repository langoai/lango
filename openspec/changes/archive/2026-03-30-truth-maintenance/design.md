## Approach

Add a TruthMaintainer layer between OntologyService facade and graph.Store. All fact assertions go through AssertFact which: validates predicates, sets temporal metadata, detects cardinality-based conflicts (OneToOne/ManyToOne), auto-resolves by source precedence, or creates open conflict records. RetractFact sets ValidTo (soft delete). FactsAt filters by temporal validity window.

## Key Decisions

- **Temporal metadata in existing Metadata map** — no BoltDB key structure change, backward-compatible with legacy triples (no temporal fields = always valid)
- **Conflict persistence in SQLite (Ent)** — not BoltDB, enabling structured queries on conflict status
- **Auto-resolve audit trail** — even auto-resolved conflicts leave a `status: auto_resolved` record for provenance tracking
- **Metadata before conflict detection** — temporal fields set before toCandidate() to avoid empty source/confidence in conflict records
- **ValidateTriple first** — AssertFact calls svc.ValidateTriple before anything else to prevent unknown predicates from bypassing registry validation
- **isCurrentlyValid checks both ValidFrom and ValidTo** — future-dated facts don't trigger present-time conflicts
- **ManyToOne = conflict-eligible** — ManyToOne means one object per subject-predicate, same constraint as OneToOne from subject side

## Dependencies

- Change 1-1 (OntologyQueryService + Registry) ✅
- Change 1-2 (Graph ABI Migration) ✅

## Files

### New
- `internal/ontology/truth.go` — TruthMaintainer interface + types + implementation
- `internal/ontology/truth_ent.go` — ConflictStore (Ent-backed CRUD)
- `internal/ontology/truth_test.go` — 17 tests
- `internal/ent/schema/ontology_conflict.go` — Ent schema

### Modified
- `internal/ontology/types.go` — MetaValidFrom etc. 6 constants + SourcePrecedence map
- `internal/ontology/service.go` — 6 new methods on OntologyService + ServiceImpl delegation
- `internal/app/wiring_ontology.go` — ConflictStore + TruthMaintainer creation
