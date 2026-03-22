## Why

Agent Memory uses an in-memory map-based store (`InMemoryStore`) that loses all agent memories on server restart. This prevents cross-session context retention — agents cannot accumulate domain-specific knowledge across conversations. Additionally, the `agentNameOrDefault()` function falls back to "default" because neither the root agent nor sub-agent tool adaptation paths inject the agent name into context, making "per-agent" storage ineffective.

## What Changes

- Add Ent schema for `AgentMemory` entity with persistent SQLite storage
- Implement `EntStore` backing all 11 `Store` interface methods with Ent queries
- Change `ToolAdapter` signature in orchestration to accept `agentName`, enabling per-agent memory isolation
- Update root agent tool adaptation to inject `"lango-agent"` name into context
- Wire `EntStore` into app modules replacing `InMemoryStore`
- Update README, CLI help text, and docs to reflect persistent storage

## Capabilities

### New Capabilities

- `agent-memory-ent-store`: Ent-backed persistent storage for agent memory with per-agent isolation, scope-based search (instance + global), and cross-session retention

### Modified Capabilities

- `agent-memory`: Store implementation changed from in-memory to Ent-backed persistent. ToolAdapter signature updated to propagate agent names. Type scope deferred (instance + global only).

## Impact

- **Code**: `internal/agentmemory/` (new ent_store.go), `internal/ent/schema/` (new agent_memory.go), `internal/orchestration/orchestrator.go` (ToolAdapter signature), `internal/app/wiring.go` (root agent + orchestrator config), `internal/app/modules.go` (wiring)
- **APIs**: `orchestration.ToolAdapter` type signature changes from `func(*agent.Tool) (tool.Tool, error)` to `func(*agent.Tool, string) (tool.Tool, error)` — internal only, no external API change
- **Dependencies**: No new dependencies. Uses existing Ent framework.
- **Systems**: Database schema auto-migrated on startup. No CLI/TUI behavioral changes (help text only).
