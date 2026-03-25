## ADDED Requirements

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

### Requirement: RecoveryPolicy decides post-failure action
The system SHALL provide a `RecoveryPolicy` that evaluates failures and returns a `RecoveryAction`: `RecoveryRetry`, `RecoveryRetryWithHint`, `RecoveryDirectAnswer`, or `RecoveryEscalate`. `RecoveryRetryWithHint` SHALL add a prompt hint requesting the root orchestrator to try a different specialist — it is NOT per-agent direct execution.

#### Scenario: Retry on transient error
- **WHEN** inner executor fails with a transient error (e.g., provider rate limit) and retry count < `recovery.maxRetries`
- **THEN** RecoveryPolicy SHALL return `RecoveryRetry`

#### Scenario: Hint retry on tool churn
- **WHEN** inner executor fails with `ErrToolChurn` and retry count < `recovery.maxRetries`
- **THEN** RecoveryPolicy SHALL return `RecoveryRetryWithHint` with failed agent in exclude list

#### Scenario: Direct answer on partial result
- **WHEN** inner executor fails but produced partial text and retry budget is exhausted
- **THEN** RecoveryPolicy SHALL return `RecoveryDirectAnswer`

#### Scenario: Escalate on unrecoverable error
- **WHEN** inner executor fails with unrecoverable error or all retries exhausted with no partial result
- **THEN** RecoveryPolicy SHALL return `RecoveryEscalate`

### Requirement: Orchestration config schema
The system SHALL support the following config keys under `agent.orchestration`:
- `mode`: `"classic"` (default) or `"structured"`
- `circuitBreaker.failureThreshold`: int (default: 3)
- `circuitBreaker.resetTimeout`: duration (default: 30s)
- `budget.toolCallLimit`: int (default: 50)
- `budget.delegationLimit`: int (default: 15)
- `budget.alertThreshold`: float64 (default: 0.8)
- `recovery.maxRetries`: int (default: 2)
- `recovery.circuitBreakerCooldown`: duration (default: 5m)

#### Scenario: Default config preserves classic mode
- **WHEN** no `agent.orchestration` section exists in config
- **THEN** the system SHALL use classic mode with no CoordinatingExecutor wrapper

### Requirement: Orchestrator direct-tool guard exempts transfer_to_agent
The orchestrator direct-tool guard (`agent.go` Run loop) SHALL exempt pure `transfer_to_agent` FunctionCall events. "Pure" means ALL FunctionCalls in the event are `transfer_to_agent`; mixed events (transfer_to_agent + real tool) SHALL still be blocked. This is required because ADK yields the model-response event (with FunctionCall) before promoting it to `Actions.TransferToAgent` in a subsequent event.

#### Scenario: Pure transfer_to_agent passes guard
- **WHEN** the orchestrator emits an event with only `transfer_to_agent` FunctionCall(s)
- **AND** `Actions.TransferToAgent` is not yet set (ADK 2-phase)
- **THEN** the guard SHALL NOT terminate the run with "orchestrator emitted direct tool call"

#### Scenario: Mixed transfer + tool blocked
- **WHEN** the orchestrator emits an event with `transfer_to_agent` AND other FunctionCalls
- **THEN** the guard SHALL terminate the run as before

### Requirement: Recovery escalates on orchestrator guard violation
RecoveryPolicy SHALL classify `CauseOrchestratorDirectTool` errors as `RecoveryEscalate` (not `RecoveryRetry`). Same-input retry cannot resolve a guard violation and would cause infinite retry loops.

#### Scenario: Guard violation escalates immediately
- **WHEN** inner executor fails with `ErrToolError` and `CauseClass == "orchestrator_direct_tool_call"`
- **THEN** RecoveryPolicy SHALL return `RecoveryEscalate`

### Requirement: Recovery diagnostic logging
Recovery event logging SHALL include `error_code` and `cause_class` from `AgentError` when available, to support root-cause analysis. Logging SHALL NOT include tool input/output payloads.

#### Scenario: AgentError recovery log includes classification
- **WHEN** recovery is triggered by an `AgentError`
- **THEN** the log entry SHALL include `error_code`, `cause_class`, `agent`, `action`, and `retry` fields
