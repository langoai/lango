## Context

The Lango agent framework uses Google ADK with Gemini as one of its LLM providers. Gemini enforces strict message turn-ordering rules that other providers (OpenAI, Anthropic) do not. Session history can become malformed through streaming partial events, token budget truncation splitting FunctionCall/Response pairs, and multi-agent delegation creating consecutive model turns. Separately, the default 25-turn limit is too low for multi-agent workflows where delegation routing consumes turns without productive tool work.

## Goals / Non-Goals

**Goals:**
- Eliminate Gemini INVALID_ARGUMENT errors from turn-order violations
- Prevent premature turn limit exhaustion in multi-agent mode
- Add observability for turn limit consumption (80% warning)
- Maintain backward compatibility for single-agent mode

**Non-Goals:**
- Refactoring the provider interface or adding provider-level retry logic
- Changing the fundamental turn-counting paradigm (still counts FunctionCall events)
- Implementing Gemini-specific content validation beyond turn ordering

## Decisions

### D1: 5-step sanitization pipeline in gemini/sanitize.go
Pipeline: (1) drop orphaned FunctionResponses → (2) merge consecutive roles → (3) prepend user if starts with model → (4) ensure FunctionCall/FunctionResponse pairs → (5) final merge pass.

**Rationale**: Each step addresses a distinct Gemini API invariant. Running merge twice (steps 2 and 5) handles synthetic user turns inserted by step 4 that may be adjacent to existing user turns. Alternatives considered: single-pass validation (rejected — too complex and brittle), modifying EventsAdapter to produce clean sequences (rejected — would affect all providers).

### D2: Defense-in-depth role merging in EventsAdapter.All()
Consecutive same-role events are merged using a pending-event buffer pattern before yielding.

**Rationale**: Primary defense is at the Gemini provider level. EventsAdapter merging prevents malformed sequences from reaching any provider, reducing the probability of turn-order errors across the system. Two independent defenses are better than one.

### D3: Delegation events excluded from turn counting
`isDelegationEvent()` checks `event.Actions.TransferToAgent != ""` and skips counting.

**Rationale**: Delegation is routing overhead, not productive tool work. In a 7-agent hierarchy, 4-6 delegation transfers per request are normal and should not consume the turn budget.

### D4: Graceful wrap-up turn
One extra turn is granted after the limit is reached before hard stop. A `wrapUpGranted` flag prevents infinite extensions.

**Rationale**: Abrupt mid-thought interruption produces poor UX. One extra turn allows the agent to finalize its response. Only one wrap-up turn is granted — no risk of unbounded extension.

### D5: Multi-agent default 50 turns
When `agent.multiAgent` is true and no explicit `MaxTurns` is configured, the default is 50 instead of 25.

**Rationale**: Multi-agent mode has inherent overhead from delegation routing. After excluding delegations, actual tool calls across 3-4 sub-agents easily reach 25. The 50-turn default provides sufficient headroom.

## Risks / Trade-offs

- [O(n) sanitization pass on every Gemini API call] → Acceptable overhead; contents are typically < 100 entries
- [Synthetic "[continue]" user turn may affect Gemini response quality] → Minimal impact; only inserted when history starts with model turn (rare edge case after truncation)
- [Synthetic FunctionResponse with "[no response available]" is a data loss marker] → Only inserted for orphaned FunctionCalls, which represent already-broken state
- [EventsAdapter merging changes Len() semantics] → Len() now reflects post-merge count via cached events, consistent with All()
