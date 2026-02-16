## Context

Multi-agent orchestration uses an ADK runner with a root agent named `lango-orchestrator`. However, `EventsAdapter.All()` hardcodes all assistant message authors as `lango-agent`. This mismatch causes the ADK runner to log "Event from unknown agent: lango-agent" on every event replay, triggering unnecessary agent graph traversal. Combined with unconditional sub-agent creation and no short-circuit for simple queries, a simple "hello" message generates 18+ events.

## Goals / Non-Goals

**Goals:**
- Eliminate "Event from unknown agent" warnings by correctly mapping event authors
- Persist agent identity in message history for multi-agent session continuity
- Reduce event count for simple queries from 18+ to ~2-3 via orchestrator short-circuit
- Only create sub-agents that have tools assigned (avoid empty delegations)
- Add delegation round limit as a safety guardrail
- Fix A2A remote agent wiring order so remote agents appear in the agent tree

**Non-Goals:**
- Programmatic delegation limiting (ADK v0.4.0 does not support it; prompt-based guardrail only)
- Changing the ADK runner's event processing logic
- Modifying the single-agent (`multiAgent: false`) code path

## Decisions

### D1: Pass rootAgentName through SessionAdapter chain
**Decision**: Thread `rootAgentName` from `NewSessionServiceAdapter` → `NewSessionAdapter` → `EventsAdapter`, using it as fallback when `msg.Author` is empty.
**Rationale**: Preserves backward compatibility (old messages without Author still work) while correctly identifying the root agent for new events. Alternative of patching EventsAdapter to detect agent name at runtime was rejected as fragile.

### D2: Store Author field in session.Message and ent schema
**Decision**: Add `Author string` to `session.Message` and a corresponding `author` column to ent Message schema (optional, default empty).
**Rationale**: Persisting the author allows accurate event replay across sessions. Without it, restarted sessions lose multi-agent routing context. The field is optional to maintain backward compatibility with existing data.

### D3: Conditional sub-agent creation based on tool partition
**Decision**: Only create executor/researcher/memory-manager sub-agents when `PartitionTools` assigns tools to their role. Planner is always created (LLM-only).
**Rationale**: Empty sub-agents waste an LLM call when the orchestrator delegates to them. Dynamic instruction generation ensures the orchestrator only knows about agents that actually exist.

### D4: Prompt-based short-circuit and delegation limit
**Decision**: Modify orchestrator instruction to allow direct responses for simple queries and state a max delegation round limit.
**Rationale**: ADK v0.4.0 lacks programmatic delegation control. Prompt engineering is the only mechanism available. Low risk since it only modifies prompt text.

### D5: Move A2A loading before BuildAgentTree
**Decision**: Reorder `wiring.go` to load remote A2A agents into `orchCfg.RemoteAgents` before calling `BuildAgentTree()`.
**Rationale**: The previous order set `RemoteAgents` after tree construction, so they were never included. Simple line reorder fix.

## Risks / Trade-offs

- **[Prompt-based guardrails are soft limits]** → The orchestrator may still occasionally delegate simple queries. Acceptable trade-off given ADK v0.4.0 constraints.
- **[Schema migration required]** → Adding `author` column requires `go generate ./internal/ent`. The column is optional with empty default, so existing data is unaffected.
- **[Fewer sub-agents may reduce capability]** → If a tool is miscategorized by prefix, its sub-agent won't be created. Mitigated by the existing default-to-executor fallback in PartitionTools.
