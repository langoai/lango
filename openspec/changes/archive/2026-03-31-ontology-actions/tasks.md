## 1. Types

- [x] 1.1 Add ActionStatus, ActionEffects, FactEffect, FactRetraction, PropertyEffect, ActionResult, ActionSummary, ActionLogEntry types to `internal/ontology/types.go`

## 2. Ent Schema

- [x] 2.1 Create `internal/ent/schema/action_log.go` — ActionLog Ent schema
- [x] 2.2 Run `go generate ./internal/ent` to generate Ent client code

## 3. ActionLog Store

- [x] 3.1 Create `internal/ontology/action_log.go` — ActionLogStore (Ent-backed CRUD)

## 4. Action Registry and Executor

- [x] 4.1 Create `internal/ontology/action.go` — ActionType, ActionRegistry, ActionExecutor, built-in actions (link_entities, set_entity_status)
- [x] 4.2 Create `internal/ontology/action_test.go` — tests for registry, executor lifecycle, ACL, built-in actions

## 5. Service Integration

- [x] 5.1 Add executor field, SetActionExecutor setter, and 4 new methods to `internal/ontology/service.go`
- [x] 5.2 Update `internal/ontology/tools.go` — change BuildTools signature, add dynamic tool generation and ontology_list_actions

## 6. Wiring and Callers

- [x] 6.1 Update `internal/app/wiring_ontology.go` — create ActionRegistry, register built-in actions, create ActionExecutor, inject
- [x] 6.2 Update `internal/app/modules.go` — BuildTools(svc) → BuildTools(svc, actionRegistry)
- [x] 6.3 Update `internal/ontology/tools_test.go` — BuildTools call signature

## 7. Downstream

- [x] 7.1 Update `prompts/agents/ontologist/IDENTITY.md` — add action tool descriptions
- [x] 7.2 Build and test: `go build -tags fts5 ./...` and `go test -tags fts5 ./internal/ontology/... -v`
