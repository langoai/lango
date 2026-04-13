## Context

Agent Memory currently uses `agentmemory.NewInMemoryStore()` (map-based, thread-safe). The Store interface has 11 methods including 2-phase scope search (instance → global). Agent name propagation is broken: root uses `AdaptToolWithTimeout` (no name), sub-agents use `adk.AdaptTool` via `orchestration.ToolAdapter` (no name). Type scope exists in the interface doc but has no real implementation.

## Goals / Non-Goals

**Goals:**
- Persistent agent memory via Ent/SQLite backing store
- Per-agent isolation through agent name propagation in both root and sub-agent paths
- Preserve existing Store interface and InMemoryStore (for tests)
- Match InMemoryStore's SearchWithContext semantics exactly

**Non-Goals:**
- Type scope implementation (requires agent_type model — deferred)
- CLI read path connecting to EntStore (scope control — deferred)
- Migration from existing in-memory data (ephemeral by nature)

## Decisions

### D1: ToolAdapter signature change over closure-in-buildSubAgent
**Decision**: Change `ToolAdapter` from `func(*agent.Tool) (tool.Tool, error)` to `func(*agent.Tool, string) (tool.Tool, error)` where the string is agentName. Inject timeout via closure at wiring site.
**Rationale**: Preserves dependency inversion — orchestration never imports adk. The closure in wiring captures timeout, while orchestrator passes spec.Name. Alternative (closure inside buildSubAgent) would require orchestration to import adk directly.

### D2: Root agent name hardcoded as "lango-agent"
**Decision**: Use `"lango-agent"` as the root agent name for tool context injection.
**Rationale**: This matches the hardcoded name in ADK agent creation. Using `cfg.Agent.Name` would require plumbing a new config field with no current value.

### D3: SearchWithContext merge order
**Decision**: Phase 1 results (agent's all entries, sorted) first, then Phase 2 results (other agents' global, sorted), then global limit.
**Rationale**: Matches InMemoryStore behavior where agent's own entries appear before global entries from other agents.

### D4: Instance + Global scope only
**Decision**: Schema enum includes only `instance` and `global`. Type scope skipped entirely.
**Rationale**: No agent_type concept exists. Adding it requires registry integration that's out of scope.

## Risks / Trade-offs

- **[Risk] ToolAdapter signature is a breaking internal change** → Mitigated by updating all callsites (orchestrator_test.go ~10 places, wiring.go 1 place). No external API impact.
- **[Risk] Ent auto-migration adds table on first boot** → SQLite handles this gracefully. No data loss risk.
- **[Trade-off] Keeping InMemoryStore** → Increases code but provides fast test backend without database setup.
