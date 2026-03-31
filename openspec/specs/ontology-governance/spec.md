## ADDED Requirements

### Requirement: Schema lifecycle FSM
The system SHALL define 5 SchemaStatus values: `proposed`, `quarantined`, `shadow`, `active`, `deprecated`. Valid transitions SHALL be: proposed→shadow, proposed→quarantined, shadow→active, shadow→quarantined, quarantined→proposed, active→deprecated.

#### Scenario: Valid transition proposed to shadow
- **WHEN** a type with status `proposed` is promoted to `shadow`
- **THEN** the transition SHALL succeed

#### Scenario: Invalid transition proposed to active
- **WHEN** a type with status `proposed` is promoted directly to `active`
- **THEN** the transition SHALL return an error (must go through shadow first)

#### Scenario: Deprecated is terminal
- **WHEN** a type with status `deprecated` is promoted to any other status
- **THEN** the transition SHALL return an error

#### Scenario: Quarantined can re-propose
- **WHEN** a type with status `quarantined` is transitioned to `proposed`
- **THEN** the transition SHALL succeed

### Requirement: Governance-forced proposed status
When GovernanceEngine is enabled, `RegisterType` and `RegisterPredicate` SHALL force the status to `proposed` regardless of input value and SHALL check rate limits before registration.

#### Scenario: RegisterType with governance forces proposed
- **WHEN** governance is enabled and `RegisterType` is called with status `active`
- **THEN** the type SHALL be registered with status `proposed`

#### Scenario: RegisterType without governance preserves status
- **WHEN** governance is nil and `RegisterType` is called with status `active`
- **THEN** the type SHALL be registered with status `active`

#### Scenario: Rate limit exceeded
- **WHEN** governance is enabled and the daily proposal count exceeds `MaxNewPerDay`
- **THEN** `RegisterType` SHALL return an error without registering

### Requirement: Rate limiting
The GovernanceEngine SHALL enforce a combined daily limit on new type and predicate proposals. The limit is configurable via `GovernancePolicy.MaxNewPerDay`.

#### Scenario: Within daily limit
- **WHEN** fewer than MaxNewPerDay proposals have been made today
- **THEN** registration SHALL succeed

#### Scenario: Exceeds daily limit
- **WHEN** MaxNewPerDay proposals have already been made today
- **THEN** registration SHALL return an error

### Requirement: SeedDefaults bypass
Seed types and predicates SHALL be registered with `active` status because `SeedDefaults` runs before `SetGovernanceEngine` in the wiring sequence.

#### Scenario: Seed types are active
- **WHEN** `SeedDefaults` runs before governance is enabled
- **THEN** all seed types and predicates SHALL have status `active`

### Requirement: Shadow predicates in validation cache
`refreshPredicateCache` SHALL include predicates with status `shadow` in addition to `active`. Predicates with status `proposed` or `quarantined` SHALL NOT be included.

#### Scenario: Shadow predicate passes validation
- **WHEN** a predicate has status `shadow` and `PredicateValidator` is called
- **THEN** the predicate SHALL pass validation

#### Scenario: Proposed predicate fails validation
- **WHEN** a predicate has status `proposed` and `PredicateValidator` is called
- **THEN** the predicate SHALL fail validation

### Requirement: Schema health reporting
The system SHALL provide a `SchemaHealth` method returning status counts for both types and predicates grouped by SchemaStatus.

#### Scenario: Health report includes all statuses
- **WHEN** `SchemaHealth` is called
- **THEN** the report SHALL include counts for each status (proposed, quarantined, shadow, active, deprecated) for both types and predicates

### Requirement: OntologyService governance methods
OntologyService interface SHALL be extended with 4 methods: `PromoteType(ctx, typeName, targetStatus, reason)`, `PromotePredicate(ctx, predName, targetStatus, reason)`, `SchemaHealth(ctx)`, `TypeUsage(ctx, typeName)`. `PromoteType` and `PromotePredicate` SHALL use `Registry.UpdateTypeStatus`/`UpdatePredicateStatus` (not `RegisterType`/`RegisterPredicate`) to update existing schema element status. `PromotePredicate` SHALL call `refreshPredicateCache()` after status change.

#### Scenario: PromoteType validates FSM
- **WHEN** `PromoteType` is called with an invalid transition
- **THEN** it SHALL return an error

#### Scenario: PromoteType applies valid transition
- **WHEN** `PromoteType` is called with a valid transition (proposed→shadow)
- **THEN** the type's status SHALL be updated
