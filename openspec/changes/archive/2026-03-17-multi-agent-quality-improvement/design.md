## Context

The multi-agent orchestration system routes user requests through an orchestrator agent to specialized sub-agents (operator, navigator, vault, librarian, automator, planner, chronicler). Testing revealed:
1. Orchestrator prompt tells the model it has no tools, then instructs it to "handle directly" and use `builtin_health` — contradiction
2. Turn budget cuts off mechanically regardless of task complexity
3. Single-word keywords (e.g., "search", "find", "run") overlap across agents
4. Sub-agent IDENTITY.md files use deprecated `[REJECT]` text patterns instead of `transfer_to_agent` ADK calls
5. Compressed tool output (`_meta.compressed`) not surfaced to sub-agents

## Goals / Non-Goals

**Goals:**
- Eliminate prompt contradictions that cause the orchestrator to attempt direct tool use
- Make routing deterministic for common keyword-overlap cases via disambiguation rules
- Prevent premature termination of complex multi-agent tasks
- Ensure sub-agents can retrieve compressed output via `tool_output_get`
- Synchronize all prompt override files with agentSpec escalation protocol

**Non-Goals:**
- Session-level metadata persistence (deferred — delegation events already in history)
- Programmatic routing (still LLM-based, but with better signals)
- New agent types or capability additions
- Changes to the ADK runner or session service

## Decisions

### D1: Dynamic Budget Expansion via Delegation Heuristics
Expand `maxTurns` by 50% (×1.5) when multi-agent complexity is detected. Trigger conditions (OR):
- Planner agent involved (intentional complex decomposition)
- 3+ total delegations (high agent back-and-forth)
- 2+ unique non-orchestrator agents (cross-domain task)

**Rationale**: Static budget penalizes legitimately complex tasks. Heuristic-based expansion avoids config complexity while covering 90%+ of multi-step scenarios. 50% expansion is conservative enough to prevent runaway loops.

**Alternative considered**: Per-task budget from planner output. Rejected because planner is optional and would require structured output parsing.

### D2: Compound Keywords + Disambiguation Fields
Replace ambiguous single-word keywords with multi-word compound phrases. Add `ExampleRequests` (3-5 concrete examples) and `Disambiguation` (negative routing hints) fields to `AgentSpec`.

**Rationale**: Single-word overlap (e.g., "search" matches both librarian and navigator) is the #1 misrouting cause. Compound keywords reduce false matches. Disambiguation provides explicit tie-breaking rules. Example requests give the LLM concrete pattern matching.

**Alternative considered**: Weighted keyword scoring. Rejected — adds complexity without clear benefit over compound keywords + disambiguation.

### D3: Universal tool_output_ Distribution
Collect `tool_output_` prefixed tools separately during partitioning, then distribute to all non-empty agent tool sets (excluding planner).

**Rationale**: Output manager tools are cross-cutting — any agent may receive compressed output. Distributing only to agents with existing tools prevents creating empty agents just for output retrieval.

### D4: Tool Count Instead of Tool Names in Routing Table
Replace individual tool name listing with tool count in the orchestrator routing table.

**Rationale**: Models over-fit on tool name strings (e.g., routing to librarian because it has `search_web` even when the request is about web browsing). Capability-based and example-based routing is more robust.

### D5: Multi-Tier Wrap-Up Budget
Default: 1 wrap-up turn. When budget is expanded (D1): 3 wrap-up turns. This gives complex tasks time to synthesize results.

**Rationale**: Single wrap-up turn is sufficient for simple tasks but inadequate when multiple agent results need consolidation.

## Risks / Trade-offs

- [Budget expansion may allow more API calls] → Capped at 1.5x, single expansion only, logged for monitoring
- [Compound keywords reduce recall for novel phrasings] → Capabilities list and Example Requests provide fallback matching
- [Removing tool names from routing table may reduce precision for edge cases] → Tool count still signals agent capability scope; capabilities list covers semantic matching
- [IDENTITY.md overrides may drift from agentSpecs again] → Tests verify escalation protocol presence in both sources
