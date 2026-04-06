## MODIFIED Requirements

### Requirement: Runtime bridge forwards tool and thinking events
The `enrichRequest` function SHALL wire `OnToolCall`, `OnToolResult`, `OnThinking`, `OnDelegation`, and `OnBudgetWarning` callbacks from `turnrunner.Request` to corresponding Bubble Tea messages via the `msgSender` interface.

#### Scenario: Delegation callback wired
- **WHEN** `enrichRequest` is called with a non-nil sender
- **THEN** `req.OnDelegation` SHALL be set to send a `DelegationMsg` with From, To, and Reason fields

#### Scenario: Budget warning callback wired
- **WHEN** `enrichRequest` is called with a non-nil sender
- **THEN** `req.OnBudgetWarning` SHALL be set to send a `BudgetWarningMsg` with Used and Max fields

## ADDED Requirements

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
