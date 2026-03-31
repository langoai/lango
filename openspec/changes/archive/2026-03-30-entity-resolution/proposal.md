## Why

Different representations of the same entity (e.g., `error:timeout` and `error:api_timeout`) create duplicate nodes in the graph with no linkage. Graph traversal and confidence propagation double-count, and queries miss related triples. Entity resolution provides canonical ID management, alias tracking, and merge/split operations to unify entities.

## What Changes

- Implement `EntityResolver` interface with `Resolve`, `RegisterAlias`, `DeclareSameAs`, `Merge`, `Split`, `Aliases`
- Add `EntityAlias` Ent schema for alias persistence (SQLite)
- Upgrade `StoreTriple` to canonicalize subject/object before storage (write path)
- Add `QueryTriples` for read-path canonicalization
- Merge operation: snapshot → replicate → retract → alias (safe ordering to avoid mid-merge canonicalization)
- Extend `OntologyService` interface with 6 entity resolution methods

## Capabilities

### New Capabilities
- `entity-resolution`: Entity alias management, canonicalization, merge/split operations

### Modified Capabilities
- `ontology-registry`: OntologyService interface extended with Resolve, DeclareSameAs, MergeEntities, SplitEntity, Aliases, QueryTriples

## Impact

- `internal/ontology/` — new resolution.go, resolution_ent.go; service.go interface extension
- `internal/ent/schema/` — new entity_alias.go schema + Ent codegen
- `internal/app/wiring_ontology.go` — AliasStore + EntityResolver initialization
- StoreTriple now canonicalizes subject/object via Resolve before AddTriple
- graph.Store interface: **no change**, BoltStore: **no change**
