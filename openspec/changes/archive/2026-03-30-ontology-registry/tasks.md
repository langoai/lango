## 1. Types and Interfaces

- [x] 1.1 Create `internal/ontology/types.go` — ObjectType, PredicateDefinition, PropertyDef, PropertyType, Cardinality, SchemaStatus, Constraint structs and constants
- [x] 1.2 Create `internal/ontology/service.go` — OntologyService interface with all method signatures (GetType, ListTypes, GetPredicate, ListPredicates, RegisterType, RegisterPredicate, DeprecateType, DeprecatePredicate, ValidateTriple, SchemaVersion, StoreTriple, PredicateValidator)
- [x] 1.3 Create `internal/ontology/registry.go` — Registry interface (internal, used only by ServiceImpl) with CRUD methods for types and predicates

## 2. Ent Schemas and Code Generation

- [x] 2.1 Create `internal/ent/schema/ontology_type.go` — OntologyType entity with fields: id (UUID), name (unique, not empty), description, properties (JSON), extends (optional), status (enum: active, deprecated), version (default 1), created_at (immutable), updated_at
- [x] 2.2 Create `internal/ent/schema/ontology_predicate.go` — OntologyPredicate entity with fields: id (UUID), name (unique, not empty), description, source_types (JSON), target_types (JSON), cardinality (enum: one_to_one, one_to_many, many_to_one, many_to_many), inverse (optional), status (enum: active, deprecated), version (default 1), created_at (immutable), updated_at
- [x] 2.3 Run `go generate ./internal/ent` and verify generated code compiles

## 3. Registry Implementation

- [x] 3.1 Create `internal/ontology/registry_ent.go` — EntRegistry implementing Registry interface with Ent client queries (RegisterType, GetType, ListTypes, DeprecateType, RegisterPredicate, GetPredicate, ListPredicates, DeprecatePredicate)
- [x] 3.2 Implement duplicate-name detection: return error on RegisterType/RegisterPredicate with existing name

## 4. Service Implementation

- [x] 4.1 Implement ServiceImpl struct with registry field, cacheMu (sync.RWMutex), activePredicates (map[string]bool)
- [x] 4.2 Implement PredicateValidator() — closure returning cached map lookup, refreshPredicateCache() for cache reload
- [x] 4.3 Implement ValidateTriple() — check predicate is active via cache (type-based validation deferred to Change 1-2)
- [x] 4.4 Implement SchemaVersion() — increment counter on register/deprecate operations
- [x] 4.5 Implement StoreTriple() — delegate to graph.Store.AddTriple (trivial pass-through in this change)
- [x] 4.6 Wire all Registry methods through ServiceImpl (GetType, ListTypes, RegisterType, etc.)

## 5. Seed Migration

- [x] 5.1 Create `internal/ontology/seed.go` — SeedDefaults(ctx, OntologyService) seeding 9 predicates with correct cardinality and source/target type constraints
- [x] 5.2 Seed 6 ObjectTypes (ErrorPattern, Tool, Fix, Session, Observation, Reflection) with PropertyDef lists matching current node ID conventions
- [x] 5.3 Implement idempotent check: skip if name already exists (no error, no duplicate)

## 6. Config and Wiring

- [x] 6.1 Create `internal/config/types_ontology.go` — OntologyConfig struct with `Enabled bool` field (default false)
- [x] 6.2 Add `Ontology OntologyConfig` field to Config struct in `internal/config/types.go`
- [x] 6.3 Create `internal/app/wiring_ontology.go` — initOntology(ctx, *ent.Client, *config.Config) creating EntRegistry, NewService, running SeedDefaults
- [x] 6.4 Call initOntology from wiring_graph.go after initGraphStore, handle disabled/error cases gracefully (log warning, don't break graph)

## 7. Tests

- [x] 7.1 Unit tests for EntRegistry: CRUD operations, duplicate detection, deprecation
- [x] 7.2 Unit tests for ServiceImpl: PredicateValidator cache behavior (register → refresh → validate), ValidateTriple, SchemaVersion increment
- [x] 7.3 Unit tests for SeedDefaults: first run creates 9+6 entries, second run is no-op
- [x] 7.4 Verify `go build -tags fts5 ./...` succeeds
- [x] 7.5 Verify existing `go test ./internal/graph/... -v` passes (no regression)
