## Why

Lango's graph store uses 9 hardcoded predicate constants (`store.go:11-21`) and untyped string node IDs with ad-hoc prefix conventions (`error:`, `tool:`, `fix:`). This prevents domain-specific relationship extension, type-safe entity validation, and structured queries. To make the agent reason more precisely — better memory, better learning, better knowledge retrieval — the graph needs a formal ontology layer: typed objects, dynamic predicates, and a single facade that all consumers use.

## What Changes

- New `internal/ontology/` package with `OntologyService` interface (facade) and `ServiceImpl`
- `Registry` interface + Ent-backed implementation for ObjectType and PredicateDefinition CRUD
- `ObjectType` struct: name, description, properties (PropertyDef), extends, status, version
- `PredicateDefinition` struct: name, description, source/target type constraints, cardinality, inverse, status, version
- Ent schemas: `ontology_type`, `ontology_predicate` (SQLite persistence)
- Seed migration: existing 9 predicates + 6 node types registered at startup (idempotent)
- `PredicateValidator()` method returning cached `func(string) bool` for hot-path injection
- `OntologyConfig` in config system (`ontology.enabled`)
- `wiring_ontology.go` initialization in app bootstrap
- No changes to existing graph code in this change (graph store integration is Change 1-2)

## Capabilities

### New Capabilities
- `ontology-registry`: ObjectType and PredicateDefinition registry with Ent-backed storage, CRUD, validation, cached predicate lookup, and seed migration for existing graph predicates/node types

### Modified Capabilities

## Impact

- New package: `internal/ontology/` (types, service, registry, seed)
- New Ent schemas: `internal/ent/schema/ontology_type.go`, `ontology_predicate.go` (requires `go generate`)
- Modified: `internal/config/types.go` (add `Ontology OntologyConfig` field)
- Modified: `internal/app/wiring_graph.go` (add `initOntology` call after graph store init)
- New: `internal/config/types_ontology.go`, `internal/app/wiring_ontology.go`
- No breaking changes to existing graph consumers — registry exists alongside current code
