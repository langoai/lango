## Context

In multi-agent mode, sub-agents that receive out-of-scope requests produce unhelpful text responses containing `[REJECT]` markers or hallucinate non-existent agent names (e.g. "assistant"). The `[REJECT]` protocol was prompt-only — no code-level enforcement existed. Users had to manually re-route requests, breaking the seamless orchestration experience.

ADK already injects `transfer_to_agent` into all sub-agents when `DisallowTransferToParent` is false (the default). This existing mechanism was unused because sub-agent prompts instructed text output instead of tool calls.

## Goals / Non-Goals

**Goals:**
- Sub-agents that cannot handle a request MUST transfer control back to the orchestrator via `transfer_to_agent` (not text)
- Orchestrator MUST re-route or answer directly when a sub-agent transfers back
- Orchestrator MUST handle simple conversational requests (greetings, general knowledge) directly without delegation
- Code-level safety net MUST catch any residual `[REJECT]` text and force re-routing

**Non-Goals:**
- Modifying `RunStreaming` — streamed text cannot be retracted; prompt-level fix covers this path
- Changing `BuildAgentTree` or `DisallowTransferToParent` — already correctly configured
- Adding new tools or modifying tool partitioning logic
- Changing the A2A remote agent integration

## Decisions

### Decision 1: `transfer_to_agent` over text-based `[REJECT]`

**Choice**: Replace all sub-agent `[REJECT]` text instructions with `transfer_to_agent` call to `lango-orchestrator`.

**Rationale**: ADK natively supports `transfer_to_agent` on all sub-agents. Using the tool guarantees immediate control transfer without relying on text parsing. The text protocol was unreliable — LLMs sometimes ignored it, produced partial matches, or added conversational text before/after the marker.

**Alternative considered**: Parse `[REJECT]` text in the orchestrator and re-route. Rejected because it adds complexity to text processing and doesn't prevent the sub-agent from emitting unhelpful text to the user in streaming mode.

### Decision 2: Three-layer defense

**Choice**: Prompt-level (Layer 1: sub-agent escalation) + Prompt-level (Layer 2: orchestrator re-routing) + Code-level (Layer 3: `[REJECT]` text safety net).

**Rationale**: LLMs are probabilistic — no single prompt change guarantees 100% compliance. The code-level safety net in `RunAndCollect` catches the residual case where a sub-agent emits `[REJECT]` text despite prompt instructions. This defense-in-depth approach minimizes user-facing failures.

### Decision 3: Orchestrator Step 0 (ASSESS) for direct response

**Choice**: Add a Step 0 to the Decision Protocol that checks if the request is simple/conversational before attempting delegation.

**Rationale**: Simple requests (greetings, weather, math) were being unnecessarily delegated, consuming round budget and increasing latency. The orchestrator is fully capable of answering these directly.

### Decision 4: `RunAndCollect`-only safety net (not `RunStreaming`)

**Choice**: Apply `[REJECT]` detection only in `RunAndCollect`, not `RunStreaming`.

**Rationale**: `RunStreaming` has already emitted partial text to the user — retraction is impossible. The prompt-level fix (Layer 1) covers streaming. `RunAndCollect` buffers the full response before returning, making retry feasible.

## Risks / Trade-offs

- **[Risk] LLM ignores escalation protocol** → Mitigated by Layer 3 code safety net detecting `[REJECT]` text and forcing re-route
- **[Risk] Retry in RunAndCollect adds latency** → Only triggers on the rare case where a sub-agent emits `[REJECT]` text; normal flow is unaffected
- **[Risk] Infinite re-routing loop** → Mitigated by single retry limit in safety net + orchestrator's round budget cap
- **[Trade-off] Streaming mode has weaker protection** → Acceptable because prompt-level fix handles the common case; full fix would require architectural changes to streaming
