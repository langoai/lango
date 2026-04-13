## Approach

Add an EntityResolver layer to OntologyService that manages entity aliases. Write path (StoreTriple) canonicalizes subject/object before storing. Read path (QueryTriples) canonicalizes the query subject. Merge uses safe ordering: snapshot triples → replicate with canonical IDs → retract originals via TruthMaintainer → register alias last.

## Key Decisions

- **Alias-last merge order** — registering alias before triple migration causes canonicalization to interfere with duplicate lookups
- **DeclareSameAs convention** — second argument treated as canonical on tie (intuitive call pattern: `DeclareSameAs(alias, canonical)`)
- **BoltStore unchanged** — all canonicalization happens in ServiceImpl facade, BoltStore remains a pure storage layer
- **Read path via QueryTriples** — separate method that canonicalizes before querying; existing graph tools (graph_traverse, graph_query) NOT modified (deferred to Stage 1.5 Ontology Tools)
- **AliasStore upsert** — Register updates existing alias if raw_id already exists

## Dependencies

- Change 1-3 (Truth Maintenance) — Merge requires RetractFact for original triple invalidation

## Files

### New
- `internal/ontology/resolution.go` — EntityResolver interface + types + implementation
- `internal/ontology/resolution_ent.go` — AliasStore (Ent-backed CRUD)
- `internal/ontology/resolution_test.go` — 11 tests
- `internal/ent/schema/entity_alias.go` — Ent schema

### Modified
- `internal/ontology/service.go` — 6 new methods on OntologyService + StoreTriple Resolve pipeline + QueryTriples
- `internal/app/wiring_ontology.go` — AliasStore + EntityResolver creation
