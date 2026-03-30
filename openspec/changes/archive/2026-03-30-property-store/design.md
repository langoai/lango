## Approach

Add an EAV (Entity-Attribute-Value) table in SQLite via Ent for per-entity property values. PropertyStore handles CRUD. OntologyService validates property names against ObjectType.Properties schema before storing, and canonicalizes entity_id via EntityResolver before all reads/writes.

## Key Decisions

- **EAV in SQLite, not BoltDB Triple.Metadata** — BoltDB has no secondary indexes on metadata values. EAV table enables indexed property queries (entity_type + property + value composite index).
- **MVP FilterOps: eq, neq, contains only** — value column is Text (string). Numeric/temporal comparison (gt/lt/gte/lte) would produce incorrect results with string ordering. Deferred to future value_type-aware CAST optimization.
- **Property validation on write** — SetEntityProperty calls GetType(entityType) and checks that property name exists in ObjectType.Properties. Unknown properties are rejected to enforce schema integrity.
- **Alias-aware read/write** — All entity_id inputs go through Resolve() before PropertyStore operations. Properties are stored under canonical IDs only.
- **GetEntity: outgoing + incoming** — Returns complete relationship picture. QueryEntities: outgoing only (incoming per-entity is too expensive for list queries).
- **PropertyStore internal to ontology** — Never exposed directly; all access through OntologyService facade.

## Dependencies

- Stage 1 complete (OntologyService, TruthMaintainer, EntityResolver)

## Files

### New
- `internal/ent/schema/entity_property.go` — EAV Ent schema
- `internal/ontology/property_store.go` — PropertyStore CRUD + Query
- `internal/ontology/property_test.go` — 12+ tests

### Modified
- `internal/ontology/types.go` — PropertyQuery, PropertyFilter, FilterOp, EntityResult types
- `internal/ontology/service.go` — 4 new methods + SetPropertyStore setter
- `internal/app/wiring_ontology.go` — PropertyStore creation + injection
