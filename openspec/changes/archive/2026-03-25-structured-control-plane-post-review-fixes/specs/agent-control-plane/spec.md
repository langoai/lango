## MODIFIED Requirements

### Requirement: CoordinatingExecutor wraps turnrunner.Executor
The system SHALL provide a `CoordinatingExecutor` type in `internal/agentrt/` that implements `turnrunner.Executor` interface (`RunStreamingDetailed`). It SHALL wrap an inner executor and apply DelegationGuard, BudgetPolicy, and RecoveryPolicy before/after delegating to the inner executor. Chunk streaming, trace hooks, and idle timeout extension SHALL be preserved through the wrapper. When the wrapper adds its own event observer, it SHALL preserve any previously installed `onEvent` hook instead of replacing it.

#### Scenario: Structured mode wraps executor
- **WHEN** config `agent.orchestration.mode` is `"structured"`
- **THEN** `initAgentRuntime()` returns a `CoordinatingExecutor` wrapping the inner `adk.Agent` executor

#### Scenario: Classic mode passes through
- **WHEN** config `agent.orchestration.mode` is `"classic"` or unset
- **THEN** `initAgentRuntime()` returns the inner executor unchanged

#### Scenario: Streaming preserved
- **WHEN** `CoordinatingExecutor.RunStreamingDetailed()` is called with an `onChunk` callback
- **THEN** the `onChunk` callback receives the same chunks as the inner executor would produce

#### Scenario: Existing event hooks preserved
- **WHEN** the inner call path already installed an `onEvent` hook before the wrapper adds its policy observer
- **THEN** both handlers SHALL run for each observed ADK event
- **AND** the wrapper SHALL NOT replace the existing handler

### Requirement: DelegationGuard observes delegations post-hoc
The system SHALL provide a `DelegationGuard` that observes delegation events via ADK event hooks and maintains per-agent circuit breaker state. It SHALL NOT make routing decisions — routing remains with the root orchestrator LLM. Circuit breaker outcomes SHALL be recorded against the delegated specialist observed during the current execution attempt, and a return transfer back to `lango-orchestrator` SHALL NOT overwrite that specialist attribution.

#### Scenario: Circuit breaker opens after threshold failures
- **WHEN** an agent has consecutive failures exceeding `circuitBreaker.failureThreshold` (default: 3)
- **THEN** the guard SHALL mark that agent's circuit as open and publish a `CircuitBreakerTrippedEvent` on the EventBus

#### Scenario: Circuit breaker resets after cooldown
- **WHEN** an agent's circuit is open and `circuitBreaker.resetTimeout` (default: 30s) has elapsed
- **THEN** the guard SHALL transition the circuit to half-open, allowing the next delegation to proceed as a probe

#### Scenario: Successful probe closes circuit
- **WHEN** an agent in half-open state completes a delegation successfully
- **THEN** the guard SHALL close the circuit and reset the failure counter

#### Scenario: Return-to-root does not steal failure attribution
- **WHEN** a specialist delegates back to `lango-orchestrator` after failing during the current attempt
- **THEN** the guard SHALL record the failure outcome against that specialist
- **AND** the root orchestrator SHALL NOT become the failure target for circuit breaker accounting

### Requirement: BudgetPolicy mirrors inner budget observationally
The system SHALL provide a `BudgetPolicy` that mirrors the inner executor's turn and delegation counts using the same counting semantics as `agent.go:350` — only events with function calls that are not delegations count as turns. It SHALL publish `BudgetAlertEvent` when thresholds are crossed. It SHALL NOT enforce limits — the inner executor's hardcoded limits remain authoritative. Mirrored counters SHALL be isolated per execution run so concurrent turns do not share or reset each other's observational state.

#### Scenario: Turn count mirrors inner semantics
- **WHEN** an ADK event contains a function call and is not a delegation event
- **THEN** `BudgetPolicy.RecordTurn()` SHALL increment the turn counter

#### Scenario: Delegation count tracks unique agents
- **WHEN** a delegation event is observed targeting agent "operator"
- **THEN** `BudgetPolicy.RecordDelegation("operator")` SHALL increment delegation counter and add "operator" to uniqueAgents set

#### Scenario: Alert at threshold
- **WHEN** turn count reaches 80% (configurable) of `budget.toolCallLimit`
- **THEN** BudgetPolicy SHALL invoke its `onAlert` callback with a `BudgetAlert`

#### Scenario: Concurrent turns keep separate mirrored counters
- **WHEN** two sessions execute concurrently through the same `CoordinatingExecutor`
- **THEN** each session SHALL maintain its own mirrored turn and delegation counters
- **AND** resetting one run's observational state SHALL NOT clear the other's counters
