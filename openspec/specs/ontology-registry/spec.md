## Purpose

Capability spec for ontology-registry. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: ObjectType registration and retrieval
The system SHALL provide an `OntologyService` interface with methods to register, retrieve, list, and deprecate ObjectType definitions. Each ObjectType SHALL have a unique name, description, list of PropertyDef, optional parent type (Extends), SchemaStatus (active/deprecated), and version number. ObjectType data SHALL be persisted in SQLite via Ent ORM.

#### Scenario: Register a new ObjectType
- **WHEN** `RegisterType(ctx, ObjectType{Name: "ErrorPattern", Properties: [...], Status: SchemaActive})` is called
- **THEN** the type is persisted in the `ontology_type` table with version 1 and the given properties

#### Scenario: Register duplicate ObjectType name
- **WHEN** `RegisterType` is called with a name that already exists
- **THEN** the system returns an error indicating the name is already registered

#### Scenario: Retrieve ObjectType by name
- **WHEN** `GetType(ctx, "ErrorPattern")` is called
- **THEN** the system returns the ObjectType with all its properties, or an error if not found

#### Scenario: List all ObjectTypes
- **WHEN** `ListTypes(ctx)` is called
- **THEN** the system returns all registered ObjectTypes regardless of status

#### Scenario: Deprecate an ObjectType
- **WHEN** `DeprecateType(ctx, "ErrorPattern")` is called
- **THEN** the type's status changes to `deprecated` and the schema version is incremented

### Requirement: PredicateDefinition registration and retrieval
The system SHALL provide methods to register, retrieve, list, and deprecate PredicateDefinition entries. Each PredicateDefinition SHALL have a unique name, description, source type constraints ([]string), target type constraints ([]string), cardinality (OneToOne/OneToMany/ManyToOne/ManyToMany), optional inverse name, SchemaStatus, and version. PredicateDefinition data SHALL be persisted in SQLite via Ent ORM.

#### Scenario: Register a new predicate
- **WHEN** `RegisterPredicate(ctx, PredicateDefinition{Name: "caused_by", SourceTypes: ["ErrorPattern"], TargetTypes: ["Tool", "ErrorPattern"], Cardinality: ManyToMany})` is called
- **THEN** the predicate is persisted in the `ontology_predicate` table and the predicate validator cache is refreshed

#### Scenario: Register duplicate predicate name
- **WHEN** `RegisterPredicate` is called with a name that already exists
- **THEN** the system returns an error indicating the name is already registered

#### Scenario: Deprecate a predicate
- **WHEN** `DeprecatePredicate(ctx, "old_predicate")` is called
- **THEN** the predicate's status changes to `deprecated` and the predicate validator cache is refreshed to exclude it

### Requirement: Cached predicate validation
The system SHALL provide a `PredicateValidator()` method that returns a `func(string) bool` closure. The closure SHALL use a cached `map[string]bool` of active predicate names, protected by `sync.RWMutex`. The cache SHALL be loaded at initialization and refreshed after any `RegisterPredicate` or `DeprecatePredicate` call. The closure SHALL NOT make database queries on each invocation.

#### Scenario: Validate known predicate
- **WHEN** `PredicateValidator()("caused_by")` is called after seeding
- **THEN** the function returns `true` without any database query

#### Scenario: Validate unknown predicate
- **WHEN** `PredicateValidator()("nonexistent_pred")` is called
- **THEN** the function returns `false`

#### Scenario: Cache refresh after registration
- **WHEN** `RegisterPredicate(ctx, PredicateDefinition{Name: "deployed_on", ...})` is called
- **AND** `PredicateValidator()("deployed_on")` is called after
- **THEN** the function returns `true`

#### Scenario: Cache refresh after deprecation
- **WHEN** `DeprecatePredicate(ctx, "old_pred")` is called
- **AND** `PredicateValidator()("old_pred")` is called after
- **THEN** the function returns `false`

### Requirement: Triple validation
The system SHALL provide a `ValidateTriple(ctx, Triple)` method that checks whether a triple's predicate is registered and active. In this change, type-based validation (SubjectType/ObjectType matching SourceTypes/TargetTypes) is NOT enforced — that is deferred to Change 1-2 when Triple gains type fields. ValidateTriple SHALL return an error if the predicate is unknown or deprecated.

#### Scenario: Validate triple with known predicate
- **WHEN** `ValidateTriple(ctx, Triple{Predicate: "caused_by"})` is called
- **THEN** validation succeeds (nil error)

#### Scenario: Validate triple with unknown predicate
- **WHEN** `ValidateTriple(ctx, Triple{Predicate: "unknown_rel"})` is called
- **THEN** validation returns an error indicating unknown predicate

### Requirement: Seed migration for existing predicates and node types
The system SHALL seed 9 existing predicates (related_to, caused_by, resolved_by, follows, similar_to, contains, in_session, reflects_on, learned_from) and 6 existing node types (ErrorPattern, Tool, Fix, Session, Observation, Reflection) at startup. Seeding SHALL be idempotent — if a predicate or type with the same name already exists, it is skipped. Each seed entry SHALL have `status: active` and `version: 1`.

#### Scenario: First startup seed
- **WHEN** the application starts with `ontology.enabled: true` and no existing ontology data
- **THEN** 9 predicates and 6 object types are created in the database

#### Scenario: Idempotent re-seed
- **WHEN** the application starts again after a previous successful seed
- **THEN** no duplicate entries are created, existing data is unchanged

#### Scenario: Seed predicate cardinality
- **WHEN** the seed runs for predicate "in_session"
- **THEN** the predicate is registered with cardinality `ManyToOne` and target types `["Session"]`

### Requirement: Schema version tracking
The system SHALL provide a `SchemaVersion(ctx)` method that returns the current schema version as an integer. The version SHALL increment whenever any ObjectType or PredicateDefinition is registered or deprecated.

#### Scenario: Initial schema version
- **WHEN** `SchemaVersion(ctx)` is called after seed migration
- **THEN** the returned version is greater than 0

#### Scenario: Version increment on registration
- **WHEN** `RegisterType(ctx, ...)` is called successfully
- **AND** `SchemaVersion(ctx)` is called before and after
- **THEN** the version after is greater than the version before

### Requirement: StoreTriple facade method
The system SHALL provide a `StoreTriple(ctx, Triple)` method on the OntologyService interface. In this change, the method SHALL delegate directly to the underlying `graph.Store.AddTriple`. In Change 1-4, this method will be extended with Resolve → Validate → Store pipeline.

#### Scenario: StoreTriple delegates to graph store
- **WHEN** `StoreTriple(ctx, triple)` is called with a valid triple
- **THEN** the triple is stored via `graph.Store.AddTriple`

### Requirement: OntologyConfig in config system
The system SHALL add an `OntologyConfig` struct with at minimum an `Enabled bool` field (default `false`). The config SHALL be nested under the `ontology` key in the config file. When `ontology.enabled` is false, no ontology initialization occurs and no schemas are seeded.

#### Scenario: Ontology disabled by default
- **WHEN** the config file has no `ontology` section
- **THEN** ontology is not initialized and no Ent migrations for ontology tables are needed

#### Scenario: Ontology enabled
- **WHEN** `ontology.enabled: true` is set in config
- **THEN** the ontology registry is initialized and seed migration runs

### Requirement: Wiring and bootstrap integration
The system SHALL initialize the ontology subsystem in `wiring_ontology.go`, called from `wiring_graph.go` after graph store initialization. If ontology is disabled or initialization fails, the graph system SHALL continue to function normally using its existing hardcoded predicate validation.

#### Scenario: Ontology init failure does not break graph
- **WHEN** ontology initialization fails (e.g., DB error)
- **THEN** the graph store continues to operate with hardcoded predicate validation
- **AND** a warning is logged
