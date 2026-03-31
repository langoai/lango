## Why

The ontology service (29 methods) is fully implemented but invisible to agents — no tools expose it. Agents cannot query types, assert facts, resolve conflicts, merge entities, or import data. This blocks the ontology from delivering value in conversations.

## What Changes

**1.5-2: Ontology Tools + Ontologist Agent**
- 10 `ontology_` prefix tools exposing OntologyService to agents (list/describe types, query/get entities, assert/retract facts, list/resolve conflicts, merge entities, facts_at)
- `ontologist` AgentSpec for tool routing (8th built-in agent)
- modules.go CatalogEntry registration
- Agent identity prompt, capabilityMap entry

**1.5-3: Ingestion Tools**
- 3 additional `ontology_` tools for bulk data ingestion (import_json, import_csv, from_mcp)
- Validation via existing SetEntityProperty/AssertFact (schema enforcement)

## Capabilities

### New Capabilities
- `ontology-tools`: 13 agent-facing tools for ontology management and data ingestion

### Modified Capabilities
- `multi-agent-orchestration`: ontologist agent added (7→8 built-in agents)
- `sub-agent-default-prompts`: ontologist IDENTITY.md added
- `agent-registry`: EmbeddedStore default agent count updated

## Impact

### Code
- `internal/ontology/tools.go` — NEW: BuildTools with 13 handlers
- `internal/orchestration/tools.go` — ontologist AgentSpec + capabilityMap
- `internal/app/modules.go` — CatalogEntry registration

### Downstream
- `prompts/agents/ontologist/IDENTITY.md` — NEW
- `README.md` — sub-agent list 7→8
- `docs/features/multi-agent.md` — agent roster update
- `openspec/specs/` — 3 spec files agent count update
- `internal/orchestration/orchestrator_test.go` — agent count assertions
