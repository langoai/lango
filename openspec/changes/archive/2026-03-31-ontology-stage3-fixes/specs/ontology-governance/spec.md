## MODIFIED Requirements

### Requirement: OntologyService governance methods
OntologyService interface SHALL be extended with 4 methods: `PromoteType`, `PromotePredicate`, `SchemaHealth`, `TypeUsage`. `PromoteType` and `PromotePredicate` SHALL use `Registry.UpdateTypeStatus`/`UpdatePredicateStatus` (not `RegisterType`/`RegisterPredicate`) to update existing schema element status. `PromotePredicate` SHALL call `refreshPredicateCache()` after status change and `version.Add(1)`.

#### Scenario: PromoteType succeeds for existing type
- **WHEN** PromoteType is called on an existing type with a valid FSM transition
- **THEN** the type status SHALL be updated (not "already exists" error)

#### Scenario: PromotePredicate refreshes cache
- **WHEN** a predicate is promoted from proposed to shadow
- **THEN** the predicate SHALL immediately pass PredicateValidator validation
