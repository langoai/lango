## 1. Ent Schema & Code Generation

- [x] 1.1 Create `internal/ent/schema/agent_memory.go` with fields: id, agent_name, scope(instance/global), kind(pattern/preference/fact/skill), key, content, confidence, use_count, tags, created_at, updated_at
- [x] 1.2 Add indexes: (agent_name, key) unique, (agent_name), (scope), (kind), (confidence)
- [x] 1.3 Run `go generate ./internal/ent/...` to generate AgentMemory entity code

## 2. EntStore Implementation

- [x] 2.1 Create `internal/agentmemory/ent_store.go` with `NewEntStore(client *ent.Client) *EntStore`
- [x] 2.2 Implement Save() with agent_name+key upsert
- [x] 2.3 Implement Get() with Where(AgentName, Key).Only(ctx)
- [x] 2.4 Implement Search() with keyword OR conditions, Kind/Tags/MinConfidence filters, sorted by confidence DESC, use_count DESC, updated_at DESC
- [x] 2.5 Implement SearchWithContext() with 2-phase: Phase 1 (agent's all entries) → Phase 2 (other agents' global), merged Phase 1 first, global limit
- [x] 2.6 Implement SearchWithContextOptions() with SearchWithContext + filters applied during collection
- [x] 2.7 Implement Delete(), IncrementUseCount(), Prune(), ListAgentNames(), ListAll(updated_at DESC)
- [x] 2.8 Add doc comment "NOTE: type scope not yet implemented" on SearchWithContext methods

## 3. Agent Name Propagation

- [x] 3.1 Change ToolAdapter signature in `orchestrator.go:18` to `func(t *agent.Tool, agentName string) (adk_tool.Tool, error)`
- [x] 3.2 Update adaptTools() in `orchestrator.go:202` to accept and pass agentName
- [x] 3.3 Update buildSubAgent() in `orchestrator.go:170` to pass spec.Name to adaptTools
- [x] 3.4 Update `wiring.go:285` root agent: `AdaptToolForAgentWithTimeout(t, "lango-agent", toolTimeout)`
- [x] 3.5 Update `wiring.go:484` orchestrator Config: closure with `AdaptToolForAgentWithTimeout(t, agentName, toolTimeout)`
- [x] 3.6 Update stubAdapter and all test adapters in `orchestrator_test.go` (~10 callsites)

## 4. Wiring

- [x] 4.1 Change `modules.go:343` from `NewInMemoryStore()` to `NewEntStore(boot.DBClient)`
- [x] 4.2 Update CatalogEntry description to "Per-agent persistent memory"

## 5. Downstream Audit

- [x] 5.1 Restore README.md "persistent memory" at 5 locations (revert "in-process" changes)
- [x] 5.2 Remove "Memory is cleared on server restart" from README.md
- [x] 5.3 Update `internal/cli/memory/agent_memory.go` help text: "persistent, retained across restarts"
- [x] 5.4 Run grep sweep to confirm zero "in-process memory" / "in-memory only" matches

## 6. Tests

- [x] 6.1 Create `internal/agentmemory/ent_store_test.go` with SQLite in-memory backend
- [x] 6.2 Test Save+Get round-trip, Upsert, Search keyword+sort, SearchWithContext phases, SearchWithContextOptions filters
- [x] 6.3 Test Delete, IncrementUseCount, Prune, ListAgentNames, ListAll ordering
- [x] 6.4 Test agent name isolation (agent A and B with same key → independent)
- [x] 6.5 Run `go test ./internal/agentmemory/... -v -count=1` — all pass
- [x] 6.6 Run `go test ./internal/orchestration/... -v -count=1` — all pass
- [x] 6.7 Run `go test ./...` — full suite passes
