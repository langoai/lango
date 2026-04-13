## Tasks

### Core Tools (1.5-2)
- [x] Create ontology/tools.go with BuildTools(svc OntologyService) signature
- [x] Implement ontology_list_types handler
- [x] Implement ontology_describe_type handler (type + relevant predicates)
- [x] Implement ontology_query_entities handler (filters array → PropertyFilter parsing)
- [x] Implement ontology_get_entity handler
- [x] Implement ontology_assert_fact handler (AssertionInput construction)
- [x] Implement ontology_retract_fact handler
- [x] Implement ontology_list_conflicts handler
- [x] Implement ontology_resolve_conflict handler (UUID parsing)
- [x] Implement ontology_merge_entities handler
- [x] Implement ontology_facts_at handler (RFC3339 time parsing)

### Ingestion Tools (1.5-3)
- [x] Implement ontology_import_json handler (JSON parse → entities loop → SetEntityProperty + AssertFact)
- [x] Implement ontology_import_csv handler (header parse → rows → entities)
- [x] Implement ontology_from_mcp handler (result_json decode → properties + fact)

### Wiring
- [x] Add ontologist AgentSpec to orchestration/tools.go agentSpecs
- [x] Add "ontology_" entry to capabilityMap in orchestration/tools.go
- [x] Replace `_ = ontologySvc` with BuildTools + CatalogEntry in modules.go

### Downstream
- [x] Create prompts/agents/ontologist/IDENTITY.md
- [x] Update README.md sub-agent list (7→8, add ontologist)
- [x] Update docs/features/multi-agent.md agent roster
- [x] Update openspec/specs/sub-agent-default-prompts/spec.md (7→8)
- [x] Update openspec/specs/multi-agent-orchestration/spec.md (7→8)
- [x] Update openspec/specs/agent-registry/spec.md (7→8)
- [x] Update orchestrator_test.go agent count + ontology prefix test

### Tests
- [x] Write tools_test.go: TestBuildTools_Count (13 tools)
- [x] Write tools_test.go: surface tool handler tests (list/describe/query/get/assert/retract/conflicts/resolve/merge/facts_at)
- [x] Write tools_test.go: ingestion tool handler tests (import_json, import_csv, from_mcp)

### Verification
- [x] go build -tags fts5 ./...
- [x] go test ./internal/ontology/... ./internal/orchestration/... -count=1
- [x] go test ./internal/graph/... ./internal/learning/... ./internal/memory/... -count=1 (regression)
