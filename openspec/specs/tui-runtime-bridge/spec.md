## ADDED Requirements

### Requirement: Runtime bridge forwards tool, thinking, delegation, and budget events
The `enrichRequest` function SHALL wire `OnToolCall`, `OnToolResult`, `OnThinking`, `OnDelegation`, and `OnBudgetWarning` callbacks from `turnrunner.Request` to corresponding Bubble Tea messages via the `msgSender` interface, without overwriting existing callbacks.

#### Scenario: Existing callbacks preserved
- **WHEN** `enrichRequest()` is called on a Request that already has `OnChunk` set
- **THEN** the existing `OnChunk` callback is NOT overwritten

#### Scenario: OnToolCall wired to ToolStartedMsg
- **WHEN** `enrichRequest()` sets `req.OnToolCall` and the callback is invoked
- **THEN** a `ToolStartedMsg` is sent to the tea.Program

#### Scenario: OnToolResult wired to ToolFinishedMsg
- **WHEN** `enrichRequest()` sets `req.OnToolResult` and the callback is invoked
- **THEN** a `ToolFinishedMsg` is sent to the tea.Program with success, duration, and output preview

#### Scenario: OnThinking wired to ThinkingStartedMsg/ThinkingFinishedMsg
- **WHEN** `enrichRequest()` sets `req.OnThinking` and the callback is invoked with `started: true`
- **THEN** a `ThinkingStartedMsg` is sent to the tea.Program

#### Scenario: Delegation callback wired
- **WHEN** `enrichRequest` is called with a non-nil sender
- **THEN** `req.OnDelegation` SHALL be set to send a `DelegationMsg` with From, To, and Reason fields

#### Scenario: Budget warning callback wired
- **WHEN** `enrichRequest` is called with a non-nil sender
- **THEN** `req.OnBudgetWarning` SHALL be set to send a `BudgetWarningMsg` with Used and Max fields

### Requirement: Bridge does not create a separate package
The bridge function SHALL reside in `internal/cli/chat/bridge.go` as a package-internal function, not in a shared UIEvent package.

#### Scenario: No uievent package exists
- **WHEN** the codebase is searched for `internal/cli/uievent/`
- **THEN** no such package exists

### Requirement: turnrunner.Request OnToolCall callback
The `turnrunner.Request` struct SHALL include an `OnToolCall func(callID, toolName string, params map[string]any)` field called when a tool invocation begins.

#### Scenario: OnToolCall fired on EventToolCall
- **WHEN** `recordEvent()` processes a `part.FunctionCall`
- **THEN** `req.OnToolCall` is invoked with callID, tool name, and parameters

#### Scenario: Nil OnToolCall is safe
- **WHEN** `req.OnToolCall` is nil and a tool call event occurs
- **THEN** the event is silently skipped without panic

### Requirement: turnrunner.Request OnToolResult callback
The `turnrunner.Request` struct SHALL include an `OnToolResult func(callID, toolName string, success bool, duration time.Duration, preview string)` field called when a tool invocation completes.

#### Scenario: OnToolResult fired on EventToolResult
- **WHEN** `recordEvent()` processes a `part.FunctionResponse`
- **THEN** `req.OnToolResult` is invoked with callID, tool name, success status, computed duration, and output preview

#### Scenario: Duration computed from callID-startedAt map
- **WHEN** a tool result arrives for a callID that was recorded in the startedAt map
- **THEN** duration equals `time.Since(startedAt)` and the map entry is cleaned up

### Requirement: turnrunner.Request OnThinking callback
The `turnrunner.Request` struct SHALL include an `OnThinking func(agentName string, started bool, summary string)` field called when thinking is detected via `genai.Part.Thought`.

#### Scenario: OnThinking fired on thought boundary
- **WHEN** `recordEvent()` encounters the first `part.Thought == true` transition (not already in thinking state)
- **THEN** `req.OnThinking` is invoked with `started: true` and the thought text; subsequent thought chunks accumulate without re-firing start

#### Scenario: OnThinking finished with accumulated summary
- **WHEN** a non-thought part arrives after one or more thought chunks
- **THEN** `req.OnThinking` is invoked with `started: false` and the full accumulated thought text as summary

### Requirement: RuntimeTracker accumulates per-turn token usage
The `RuntimeTracker` SHALL subscribe to `TokenUsageEvent` on the EventBus and accumulate token counts per turn. Events with a non-empty `SessionKey` that differs from the local session key SHALL be rejected. Events SHALL only be accumulated while `turnActive` is true.

#### Scenario: Token accumulation during active turn
- **WHEN** `StartTurn()` has been called and a `TokenUsageEvent` with matching or empty SessionKey is published
- **THEN** the token counts SHALL be accumulated in the internal snapshot

#### Scenario: Tokens ignored when turn inactive
- **WHEN** `StartTurn()` has NOT been called and a `TokenUsageEvent` is published
- **THEN** the token counts SHALL NOT be accumulated

#### Scenario: Foreign session key rejected
- **WHEN** a `TokenUsageEvent` with a non-empty SessionKey different from localSessionKey is published
- **THEN** the event SHALL be ignored

#### Scenario: FlushTurnTokens returns and resets
- **WHEN** `FlushTurnTokens()` is called
- **THEN** the accumulated snapshot SHALL be returned and internal counters reset to zero

### Requirement: RuntimeTracker forwards recovery decisions
The `RuntimeTracker` SHALL subscribe to `RecoveryDecisionEvent` on the EventBus and forward matching events as `RecoveryMsg` via the stored `msgSender`.

#### Scenario: Recovery forwarded for local session
- **WHEN** a `RecoveryDecisionEvent` with matching SessionKey is published
- **THEN** a `RecoveryMsg` SHALL be sent via the sender with CauseClass, Action, Attempt, and Backoff

#### Scenario: Foreign session recovery ignored
- **WHEN** a `RecoveryDecisionEvent` with a different SessionKey is published
- **THEN** no message SHALL be sent

### Requirement: RuntimeTracker provides turn lifecycle
The `RuntimeTracker` SHALL provide `StartTurn()`, `ResetTurn()`, `RecordDelegation(to)`, `SetActiveAgent(name)`, and `Snapshot()` methods for cockpit state management.

#### Scenario: StartTurn activates token accumulation
- **WHEN** `StartTurn()` is called
- **THEN** `Snapshot().IsRunning` SHALL return true

#### Scenario: ResetTurn clears non-token state
- **WHEN** `ResetTurn()` is called
- **THEN** delegation count, active agent, and turnActive flag SHALL be cleared (tokens are cleared by FlushTurnTokens only)

#### Scenario: SetActiveAgent updates label without counter
- **WHEN** `SetActiveAgent("lango-orchestrator")` is called
- **THEN** the active agent label SHALL update but the delegation counter SHALL NOT increment
