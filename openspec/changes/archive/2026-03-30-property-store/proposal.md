## Why

ObjectType.Properties defines schema (field names, types, constraints), but there is no per-entity instance property storage. You can't ask "which ErrorPattern has tool_name=http_client" because property values aren't persisted per entity. This blocks Stage 1.5-2 (Ontology Tools) which needs `ontology_query_entities` and `ontology_get_entity` — both require property-based filtering and entity detail retrieval.

## What Changes

- Add `EntityProperty` Ent schema (EAV model: entity_id + property + value)
- Implement `PropertyStore` with SetProperty, GetProperties, DeleteProperties, Query (eq/neq/contains filters)
- Add 4 methods to `OntologyService`: SetEntityProperty (with schema validation), GetEntityProperties, QueryEntities, GetEntity
- All read paths canonicalize entity_id via EntityResolver (alias-aware)
- All write paths validate property name against ObjectType.Properties (schema integrity)

## Capabilities

### New Capabilities
- `property-store`: Per-entity property storage (EAV), structured query with type+property filters

### Modified Capabilities
- `ontology-registry`: OntologyService extended with SetEntityProperty, GetEntityProperties, QueryEntities, GetEntity

## Impact

- `internal/ontology/` — new property_store.go; types.go extensions; service.go 4 methods
- `internal/ent/schema/` — new entity_property.go + Ent codegen
- `internal/app/wiring_ontology.go` — PropertyStore creation + injection
- No graph.Store interface change, no BoltStore change
