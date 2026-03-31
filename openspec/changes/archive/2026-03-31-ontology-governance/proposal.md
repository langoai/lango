## Why

OntologyService currently has only 2 schema states (`active`, `deprecated`). Any agent with write permission can register unlimited types/predicates instantly, with no review period, rate limiting, or lifecycle management. This creates risk of schema explosion — uncontrolled type proliferation that degrades ontology quality. Stage 2 needs governance to introduce a multi-stage lifecycle FSM with rate limiting and controlled promotion.

## What Changes

- Expand `SchemaStatus` to 5 states: `proposed`, `quarantined`, `shadow`, `active`, `deprecated`
- Add `GovernanceEngine` with FSM validation, rate limiting, and schema health reporting
- Modify `RegisterType`/`RegisterPredicate` to force `proposed` status when governance is enabled
- Expand Ent schema enums for `ontology_type` and `ontology_predicate`
- Extend `OntologyService` with 4 new methods: `PromoteType`, `PromotePredicate`, `SchemaHealth`, `TypeUsage`
- Add 4 new agent tools: `ontology_promote_type`, `ontology_promote_predicate`, `ontology_schema_health`, `ontology_type_usage`
- Update `refreshPredicateCache` to include `shadow` status predicates
- SeedDefaults bypass: governance injected after seed (wiring order guarantees)

## Capabilities

### New Capabilities
- `ontology-governance`: Schema lifecycle FSM (proposed→shadow→active→deprecated), rate limiting, schema health monitoring, type/predicate promotion tools.

### Modified Capabilities
- `ontology-tools`: 4 new governance tools added. Ontologist identity updated with governance tool descriptions.

## Impact

- `internal/ontology/` — new `governance.go`, modified `service.go` (+4 methods, RegisterType/RegisterPredicate behavior change), `types.go` (SchemaStatus expansion), `tools.go` (4 new tools)
- `internal/ent/schema/` — `ontology_type.go` and `ontology_predicate.go` status enum expansion → requires `go generate ./internal/ent`
- `internal/config/types_ontology.go` — `OntologyGovernanceConfig`
- `internal/app/wiring_ontology.go` — GovernanceEngine creation (after SeedDefaults)
- `prompts/agents/ontologist/IDENTITY.md` — governance tool descriptions
- Backward compatible: governance nil = existing behavior
