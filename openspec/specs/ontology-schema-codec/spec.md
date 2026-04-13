## Purpose

Capability spec for ontology-schema-codec. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: Slim wire types
The system SHALL define `SchemaTypeSlim` (Name, Description, Properties as `[]SchemaPropertySlim`, Extends) and `SchemaPredicateSlim` (Name, Description, SourceTypes, TargetTypes, Cardinality, Inverse) and `SchemaPropertySlim` (Name, Type, Required) in the ontology package. These types SHALL NOT contain UUID, timestamps, status, or version fields.

#### Scenario: Slim type has no local-only fields
- **WHEN** a `SchemaTypeSlim` is marshaled to JSON
- **THEN** the output SHALL NOT contain `id`, `createdAt`, `updatedAt`, `status`, or `version` keys

### Requirement: SchemaBundle format
The system SHALL define `SchemaBundle` with fields: Version (int), SchemaVersion (int), ExportedAt (time), ExportedBy (string), Types (`[]SchemaTypeSlim`), Predicates (`[]SchemaPredicateSlim`), Digest (string SHA256).

#### Scenario: Bundle uses slim types
- **WHEN** `ExportSchema` produces a SchemaBundle
- **THEN** the Types and Predicates fields SHALL contain slim wire types, not full ObjectType/PredicateDefinition

### Requirement: ExportSchema method
`OntologyService.ExportSchema(ctx)` SHALL return a SchemaBundle containing only types and predicates with status `active` or `shadow`. It SHALL require `PermRead`. The Digest field SHALL be computed from the canonical JSON of the slim types.

#### Scenario: Export includes active and shadow
- **WHEN** the ontology has types with status active, shadow, proposed, deprecated
- **THEN** ExportSchema SHALL include only active and shadow types

#### Scenario: Export digest is stable
- **WHEN** ExportSchema is called twice on the same ontology (no changes between calls)
- **THEN** both SchemaBundle.Digest values SHALL be identical

### Requirement: ImportSchema method
`OntologyService.ImportSchema(ctx, bundle, opts)` SHALL import slim types into the local ontology. It SHALL require `PermWrite`. Import SHALL convert slim types to full ObjectType/PredicateDefinition with generated UUID, current timestamps, and status determined by ImportMode. After successful import, it SHALL call `refreshPredicateCache()` if predicates were added and `version.Add(n)` where n is the total number of added types and predicates.

#### Scenario: Import shadow mode
- **WHEN** ImportSchema is called with mode `ImportShadow`
- **THEN** imported types SHALL have status `shadow`

#### Scenario: Import governed mode
- **WHEN** ImportSchema is called with mode `ImportGoverned`
- **THEN** imported types SHALL have status `proposed`

#### Scenario: Import dry run
- **WHEN** ImportSchema is called with mode `ImportDryRun`
- **THEN** ImportResult SHALL report counts but no mutations SHALL occur

#### Scenario: Import name conflict
- **WHEN** a type in the bundle has the same name as an existing local type with different properties
- **THEN** ImportResult.TypesConflicting SHALL include that type name and it SHALL NOT be imported

#### Scenario: Import skips existing identical
- **WHEN** a type in the bundle has the same name and identical properties as an existing local type
- **THEN** ImportResult.TypesSkipped SHALL be incremented

### Requirement: ComputeDigest determinism
`ComputeDigest` SHALL produce a SHA256 hash from canonical JSON (sorted keys, no whitespace) of the Types and Predicates arrays. The same logical schema SHALL always produce the same digest regardless of array ordering.

#### Scenario: Digest is order-independent
- **WHEN** two bundles have the same types in different order
- **THEN** ComputeDigest SHALL return the same digest for both

### Requirement: Full-to-slim conversion
The system SHALL provide `TypeToSlim(ObjectType) SchemaTypeSlim` and `PredicateToSlim(PredicateDefinition) SchemaPredicateSlim` conversion functions. Reverse conversion `SlimToType(SchemaTypeSlim) ObjectType` SHALL generate a new UUID and set CreatedAt/UpdatedAt to current time.

#### Scenario: Roundtrip preserves semantics
- **WHEN** a type is exported as slim then imported back
- **THEN** the imported type SHALL have the same Name, Description, Properties, Extends as the original
