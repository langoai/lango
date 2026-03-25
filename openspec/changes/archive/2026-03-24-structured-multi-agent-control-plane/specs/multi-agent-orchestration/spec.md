## ADDED Requirements

### Requirement: DelegationGuard monitors orchestrator delegations
The `DelegationGuard` SHALL observe delegation events emitted by the root orchestrator and maintain per-agent circuit breaker state. When a circuit-open agent is targeted, the guard SHALL log a warning and publish a `CircuitBreakerTrippedEvent`. The guard SHALL NOT block or redirect delegations — routing authority remains with the root orchestrator LLM.

#### Scenario: Warn on delegation to circuit-open agent
- **WHEN** root orchestrator delegates to an agent whose circuit is open
- **THEN** DelegationGuard SHALL log a warning with agent name and circuit state

### Requirement: Doctor multi-agent checks extended
The existing `MultiAgentCheck` in doctor SHALL be extended with:
- Loop frequency: count traces with `outcome=loop_detected` in last 24h via `RecentByOutcome`, warn if >3
- Timeout frequency: count traces with `outcome=timeout` in last 24h, warn if >5
- Trace store growth: `TraceCount()` vs configured maxTraces, warn if >80%
- Average turn duration: mean `ended_at - started_at` of recent successful traces, warn if >2min

#### Scenario: Loop frequency warning
- **WHEN** 5 traces with `outcome=loop_detected` exist in the last 24 hours
- **THEN** doctor SHALL emit a warning with loop count and recommendation

#### Scenario: Trace growth warning
- **WHEN** `TraceCount()` returns 8500 and maxTraces is 10000
- **THEN** doctor SHALL emit a warning that trace store is at 85% capacity

### Requirement: Gateway delegation WebSocket events
The gateway SHALL broadcast `agent.delegation` and `agent.budget_warning` WebSocket events when the corresponding TurnRunner callbacks fire.

#### Scenario: Delegation event broadcast
- **WHEN** `Request.OnDelegation` callback fires with from="orchestrator", to="operator"
- **THEN** gateway SHALL broadcast `agent.delegation` event to session clients

### Requirement: TurnRunner delegation and budget callbacks
`turnrunner.Request` SHALL support optional `OnDelegation func(from, to, reason string)` and `OnBudgetWarning func(used, max int)` callbacks. These callbacks SHALL be invoked by the turn runner when delegation events are detected in the trace recorder and when delegation count approaches the configured threshold.

#### Scenario: Callback fires on delegation
- **WHEN** the trace recorder observes a delegation event and `Request.OnDelegation` is non-nil
- **THEN** the callback SHALL be invoked with the source agent, target agent, and reason

#### Scenario: Nil callback is no-op
- **WHEN** `Request.OnDelegation` is nil and a delegation event occurs
- **THEN** no callback SHALL be invoked and execution SHALL continue normally
