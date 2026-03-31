## Why

OntologyService currently provides 29 atomic methods for individual operations (assert fact, set property, merge entities, etc.). There is no way to compose these into reusable, audited transaction units with precondition checks, compensation on failure, and execution logging. Stage 2 needs a transactional action layer so that common multi-step ontology workflows can be defined once, executed safely with ACL enforcement, and audited via structured logs.

## What Changes

- Add `ActionType` struct with Precondition/Execute/Compensate closure fields and `RequiredPerm` (in-process built-in registry, NOT persisted DSL)
- Add `ActionRegistry` for registering and looking up action types at startup
- Add `ActionExecutor` that orchestrates: ACL check → precondition → audit log → execute → compensate on failure
- Add `ActionLogStore` (Ent-backed) for structured execution records
- Add Ent schema `action_log` for persistence
- Add 2 built-in actions: `link_entities` (PermWrite), `set_entity_status` (PermWrite)
- Extend `OntologyService` interface with 4 new methods: `ExecuteAction`, `ListActions`, `GetActionLog`, `ListActionLogs`
- **BREAKING** change `BuildTools` signature: `BuildTools(svc, reg)` — dynamic tool generation from ActionRegistry
- Add `ontology_list_actions` tool and dynamic `ontology_action_{name}` tools
- Wire ActionRegistry, built-in actions, and ActionExecutor in `wiring_ontology.go`

## Capabilities

### New Capabilities
- `ontology-actions`: Transactional action execution with precondition/compensation lifecycle, ACL enforcement via direct ACLPolicy injection, structured ActionLog persistence, dynamic tool generation from action registry.

### Modified Capabilities
- `ontology-tools`: BuildTools signature changes from `BuildTools(svc)` to `BuildTools(svc, reg)`. Dynamic tools generated from ActionRegistry. New `ontology_list_actions` tool added. Ontologist identity updated with action tool descriptions.

## Impact

- `internal/ontology/` — new `action.go`, `action_log.go`, modified `service.go` (+4 methods), `tools.go` (signature change + dynamic tools), `types.go` (new types)
- `internal/ent/schema/` — new `action_log.go` schema → requires `go generate ./internal/ent`
- `internal/app/wiring_ontology.go` — ActionRegistry + ActionExecutor creation and injection
- `internal/app/modules.go` — BuildTools call site update
- `internal/ontology/tools_test.go` — BuildTools call update
- `prompts/agents/ontologist/IDENTITY.md` — action tool descriptions
