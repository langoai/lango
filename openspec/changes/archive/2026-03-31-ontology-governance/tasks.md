## 1. Types

- [x] 1.1 Add SchemaProposed, SchemaQuarantined, SchemaShadow constants and GovernancePolicy, SchemaHealthReport, TypeUsageInfo types to `internal/ontology/types.go`

## 2. Ent Schema

- [x] 2.1 Expand status enum to 5 values in `internal/ent/schema/ontology_type.go` and `internal/ent/schema/ontology_predicate.go`
- [x] 2.2 Run `go generate ./internal/ent` to regenerate Ent client code

## 3. GovernanceEngine

- [x] 3.1 Create `internal/ontology/governance.go` — GovernanceEngine, FSM validation, rate limiting, schema health
- [x] 3.2 Create `internal/ontology/governance_test.go` — FSM, rate limit, schema health tests

## 4. Service Integration

- [x] 4.1 Add governance field, SetGovernanceEngine setter, and 4 new methods to `internal/ontology/service.go`
- [x] 4.2 Modify RegisterType/RegisterPredicate to force proposed status when governance enabled
- [x] 4.3 Update refreshPredicateCache to include shadow status

## 5. Tools

- [x] 5.1 Add 4 governance tools to `internal/ontology/tools.go`

## 6. Config and Wiring

- [x] 6.1 Add OntologyGovernanceConfig to `internal/config/types_ontology.go`
- [x] 6.2 Add GovernanceEngine creation in `internal/app/wiring_ontology.go` (after SeedDefaults)

## 7. Downstream

- [x] 7.1 Update `prompts/agents/ontologist/IDENTITY.md` with governance tool descriptions
- [x] 7.2 Build and test: `go build -tags fts5 ./...` and `go test -tags fts5 ./internal/ontology/... -v`
