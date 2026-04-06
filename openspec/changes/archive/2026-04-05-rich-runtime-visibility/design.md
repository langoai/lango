## Context

Cockpit Phase 2 added channel-aware message display. However, during a turn, the user only sees "Working..." followed by tool start/finish and thinking start/finish. Delegation chains, recovery retries, budget warnings, and token consumption are invisible. The data already exists in the runtime (turnrunner callbacks, EventBus events) but has no UI surface.

## Goals / Non-Goals

**Goals:**
- Surface delegation, recovery, budget warning, and per-turn token events in the cockpit transcript
- Show live runtime state (active agent, delegation count, token usage) in the context panel
- Keep all features gracefully degraded when structured orchestration is disabled

**Non-Goals:**
- Tool intermediate progress (ADK tools execute atomically)
- Delegation tree/timeline as a separate page
- Channel-originated turn runtime details (scope: local cockpit turn only)

## Decisions

### 1. Dual data path: Callbacks vs EventBus

**Decision**: Delegation and budget warnings use turnrunner.Request callbacks (bridge.go). Recovery and token usage use EventBus subscriptions (runtimebridge.go).

**Why**: OnDelegation and OnBudgetWarning callbacks already exist on `turnrunner.Request` and fire synchronously during the local turn. Recovery events (`RecoveryDecisionEvent`) and token usage (`TokenUsageEvent`) are published to EventBus by agentrt and the model adapter respectively — they have no callback equivalent on Request.

### 2. RuntimeTracker pattern (like ChannelTracker)

**Decision**: Create `RuntimeTracker` struct in cockpit package that subscribes to EventBus, accumulates tokens, and provides Snapshot()/FlushTurnTokens()/StartTurn()/ResetTurn() lifecycle methods.

**Why**: Mirrors the established `ChannelTracker` pattern. Token events arrive per-model-call (multiple per turn), so accumulation is needed. The tracker owns the lifecycle (start/flush/reset) so cockpit.go can orchestrate timing.

### 3. Session key filtering + turnActive gating

**Decision**: Token events are filtered by `SessionKey` match (or empty, since the production publisher doesn't set it) AND by `turnActive` flag. The `turnActive` flag is set when the first content event (ToolStarted/ThinkingStarted/Chunk) arrives, and cleared on DoneMsg.

**Why**: The production `wireModelAdapterTokenUsage` publishes `TokenUsageEvent` without SessionKey. Without `turnActive` gating, tokens from channel or background turns in the same process would be incorrectly attributed to the local turn.

### 4. DoneMsg-first ordering for token summary

**Decision**: Cockpit intercepts DoneMsg, forwards it to chat child FIRST (so assistant response is appended), THEN flushes tokens and sends TurnTokenUsageMsg (so token summary appears after the response).

**Why**: Without this ordering, the token summary would appear above the assistant response in the transcript.

### 5. Orchestrator return hop handling

**Decision**: DelegationMsg with `To == "lango-orchestrator"` updates activeAgent label but does NOT increment delegation counter, matching the budget-warning logic in `turnrunner/runner.go:504`.

**Why**: Return hops are not user-visible delegations. Including them would double-count and disagree with the budget warning thresholds.

### 6. Context panel tick refresh

**Decision**: Runtime status is pushed to the context panel on every 5-second tick (alongside channel status), not just on delegation/done events.

**Why**: Token accumulation happens continuously during a turn. Without tick-based refresh, the context panel would only update on delegation events, making the runtime section stale for most of the turn.

## Risks / Trade-offs

- **[Risk] Empty SessionKey attribution**: If multiple local sessions run in the same process, tokens with empty SessionKey would be attributed to whichever turn is active → Mitigated by `turnActive` gating (only one local turn runs at a time in cockpit mode)
- **[Risk] RecoveryDecisionEvent lacks AgentName**: The context panel cannot show which agent triggered recovery → Acceptable for v1; can be added by extending the event type later
- **[Trade-off] No intermediate tool progress**: ADK tools are atomic, no streaming from within a tool → Deferred to a future framework change
