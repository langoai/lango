## Context

The orchestrator's `buildOrchestratorInstruction()` in `internal/orchestration/tools.go` includes an ASSESS step (line 663) and Delegation Rules (line 712) that list `weather` and `general knowledge` as direct-answer topics. When a user asks "tell me today's weather," the gemini-3-flash-preview model either tries to call search/browser tools directly (triggering the E003 guard) or hallucinates an answer. The existing `multi-agent-orchestration` spec explicitly includes these terms in the direct-answer scope, so a spec delta is required.

## Goals / Non-Goals

**Goals:**
- Narrow the orchestrator's direct-answer scope to topics that never require real-time data or tool access (greeting, opinion, math, small talk)
- Add an explicit function-call prohibition guard in the ASSESS block to prevent tool hallucination during direct response
- Fix regression tests to lock down the direct-answer list content

**Non-Goals:**
- Changing the recovery policy (RecoveryEscalate stays — per agent-control-plane spec)
- Token usage optimization (separate change)
- FTS5 query escaping fix (separate change, unrelated bug)

## Decisions

### Decision 1: Remove weather/general knowledge from direct-answer scope

**Choice**: Remove both terms from ASSESS step 0 and Delegation Rules #1, routing them through sub-agent delegation instead.

**Rationale**: `weather` requires real-time data (search or API). `general knowledge` is too broad — the LLM cannot distinguish between answerable-from-training questions and search-required questions. Both lead to the orchestrator attempting tool calls it cannot make.

**Alternative considered**: Narrow `general knowledge` to "factual knowledge answerable without search" — rejected because the boundary is too ambiguous for reliable LLM classification.

### Decision 2: Add MUST NOT emit function calls guard

**Choice**: Insert an explicit guard statement in the ASSESS block: "Even when responding directly, you MUST NOT emit any function calls."

**Rationale**: The existing "You do NOT have tools" instruction appears once at the top of the orchestrator prompt. LLMs with long context may miss or downweight instructions far from the decision point. Repeating the prohibition near the ASSESS step reinforces the constraint at the exact point where the LLM decides between direct response and delegation.

### Decision 3: Exclude time/date from direct-answer list

**Choice**: Do not add `time/date` to the direct-answer list.

**Rationale**: The runtime context does not inject the current date/time by default. Adding time/date to the direct-answer scope would create the same hallucination risk as weather.

## Risks / Trade-offs

- **[Risk] Increased delegation for previously direct-answered queries** → Some general knowledge questions that the LLM could have answered directly will now route through a sub-agent. Mitigation: Sub-agents handle these quickly; the additional round-trip latency is minimal compared to the current 100% failure rate.
- **[Risk] Model-specific behavior** → Prompt changes may affect different LLM models differently. Mitigation: Verified on gemini-3-flash-preview (the failing model); manual smoke testing recommended for other models.
