## Approach

Create `BuildTools(svc OntologyService) []*agent.Tool` following the graph/tools.go pattern. Each handler extracts params via toolparam helpers and delegates to the corresponding OntologyService method. Register an `ontologist` AgentSpec with Prefixes `["ontology_"]` for tool routing. Ingestion tools (import_json, import_csv, from_mcp) use existing SetEntityProperty/AssertFact for validation — no new validation code needed.

## Key Decisions

- **Single BuildTools function** — all 13 tools in one function, same file. No separate ingestion file needed since they share the same OntologyService dependency.
- **ontologist as 8th built-in agent** — Prefixes `["ontology_"]` ensures clean routing with zero overlap. Requires downstream updates to all "7 agents" references.
- **filters param as raw array** — `ontology_query_entities` filters parameter is a JSON array of objects `[{"property": "x", "op": "eq", "value": "y"}]`. Handler manually parses `[]interface{}` → `[]PropertyFilter`.
- **ontology_from_mcp uses result_json (string)** — not raw object. JSON string boundary prevents `map[string]interface{}` proliferation in handler code.
- **Validation delegation** — import tools call SetEntityProperty (schema validation) and AssertFact (predicate validation + temporal metadata). No duplicate validation in handlers.

## Dependencies

- Stage 1.5-1 PropertyStore complete ✅

## Files

### New
- `internal/ontology/tools.go` — BuildTools + 13 handlers
- `internal/ontology/tools_test.go` — 15+ tests
- `prompts/agents/ontologist/IDENTITY.md` — agent prompt

### Modified
- `internal/orchestration/tools.go` — AgentSpec + capabilityMap
- `internal/app/modules.go` — CatalogEntry
- `internal/orchestration/orchestrator_test.go` — agent count
- `README.md` — sub-agent list
- `docs/features/multi-agent.md` — agent roster
- `openspec/specs/sub-agent-default-prompts/spec.md` — 7→8
- `openspec/specs/multi-agent-orchestration/spec.md` — 7→8
- `openspec/specs/agent-registry/spec.md` — 7→8
