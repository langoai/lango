## Why

Multi-agent orchestration testing revealed systematic misrouting, premature task termination, and incomplete results. Root causes: orchestrator prompt contradictions (tool-less agent told to "handle directly"), mechanical turn budget cutoffs ignoring task complexity, ambiguous single-word routing keywords causing agent overlap, and sub-agents using deprecated `[REJECT]` patterns instead of `transfer_to_agent` escalation.

## What Changes

- Remove orchestrator prompt contradictions (Diagnostics section, "handle directly" for unmatched tools)
- Replace ambiguous single-word keywords with compound keywords for routing precision
- Add ExampleRequests and Disambiguation fields to AgentSpec for richer routing signals
- Implement dynamic turn budget expansion (1.5x) when multi-agent complexity is detected (planner involvement, 3+ delegations, or 2+ unique agents)
- Replace single wrap-up turn with configurable wrap-up budget (1 default, 3 when expanded)
- Replace "partial answer" guidance with explicit completion-first + transparency policy
- Distribute `tool_output_get` universally to all tool-bearing agents for output awareness
- Add Output Handling instructions to all non-planner sub-agent prompts
- Add Disambiguation Rules, Complexity Analysis (SIMPLE/COMPOUND/COMPLEX), and strengthened Re-Routing Protocol to orchestrator
- Replace `[REJECT]` patterns with `transfer_to_agent` escalation in all prompt override files
- Reduce tool name exposure in routing table (count instead of names) to prevent keyword over-fitting
- Add structured plan output format for planner agent

## Capabilities

### New Capabilities

### Modified Capabilities
- `multi-agent-orchestration`: Orchestrator prompt overhaul — disambiguation rules, complexity analysis phases, strengthened re-routing, output awareness, tool count instead of names
- `agent-routing`: Compound keywords, ExampleRequests, Disambiguation fields, universal tool_output_ distribution
- `agent-turn-limit`: Dynamic budget expansion based on delegation patterns, multi-tier wrap-up budget

## Impact

- `internal/orchestration/tools.go` — AgentSpec struct, agentSpecs data, PartitionTools/PartitionToolsDynamic, buildOrchestratorInstruction, routingEntry
- `internal/adk/agent.go` — Run() loop delegation tracking, budget expansion, wrap-up mechanics
- `prompts/agents/*/IDENTITY.md` (7 files) — [REJECT] → transfer_to_agent, Output Handling section
- `internal/agentregistry/defaults/*/AGENT.md` (6 files) — Output Handling section
- Tests: `orchestrator_test.go`, `agent_test.go` — new tests for universal distribution, disambiguation, budget expansion, wrap-up mechanics
