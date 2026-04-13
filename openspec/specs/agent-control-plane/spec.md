## Purpose

Capability spec for agent-control-plane. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: Exponential backoff on recovery retry
The system SHALL compute an exponential backoff delay before each recovery retry using the formula `min(baseDelay * 2^attempt, maxBackoff)` where `baseDelay` is 1 second and `maxBackoff` is 30 seconds. The backoff delay SHALL respect context cancellation — if the context is cancelled during the backoff wait, the retry SHALL be abandoned immediately.

#### Scenario: First retry waits 1 second
- **WHEN** recovery decides to retry with attempt=0
- **THEN** `ComputeBackoff(0)` SHALL return 1 second

#### Scenario: Third retry waits 4 seconds
- **WHEN** recovery decides to retry with attempt=2
- **THEN** `ComputeBackoff(2)` SHALL return 4 seconds

#### Scenario: Backoff caps at 30 seconds
- **WHEN** recovery decides to retry with attempt=10
- **THEN** `ComputeBackoff(10)` SHALL return 30 seconds (not 1024 seconds)

#### Scenario: Backoff interrupted by context cancellation
- **WHEN** the context is cancelled during the backoff sleep
- **THEN** the retry SHALL be abandoned and the context error SHALL propagate

### Requirement: Per-error-class retry limits
The system SHALL enforce per-error-class retry limits in addition to the global `maxRetries`. Each cause class SHALL have a maximum retry count. When the per-class retry count for a given cause class is exceeded, `Decide()` SHALL return `RecoveryEscalate` (or `RecoveryDirectAnswer` if partial result exists), regardless of whether the global limit has been reached. The default per-class limits SHALL be: `CauseRateLimit: 5`, `CauseTransient: 3`, `CauseMalformedToolCall: 1`, `CauseTimeout: 3`.

#### Scenario: Malformed tool call retries once then escalates
- **WHEN** a malformed tool-call error has been retried 1 time
- **THEN** `Decide()` SHALL return `RecoveryEscalate`

#### Scenario: Rate limit allows more retries than global default
- **WHEN** a rate-limit error occurs with global `maxRetries=2` and per-class limit of 5
- **THEN** the per-class limit of 5 SHALL be used as the effective retry limit

#### Scenario: Unknown cause class uses global limit
- **WHEN** a cause class is not in the per-class limit map
- **THEN** the global `maxRetries` SHALL be used as the effective retry limit

### Requirement: Malformed tool-call cause class
The system SHALL define a `CauseMalformedToolCall` cause class that maps to tool-call JSON parse errors. When `RecoveryPolicy.Decide()` encounters an `AgentError` with `CauseClass` matching `CauseFunctionCallValidation`, it SHALL classify the error as `CauseMalformedToolCall` for retry-limit purposes.

#### Scenario: Function call validation maps to malformed tool call
- **WHEN** an `AgentError` has `CauseClass == "function_call_validation"`
- **THEN** the recovery policy SHALL apply the `CauseMalformedToolCall` retry limit (1 retry)

### Requirement: Provider failure tracking in circuit breaker
The `DelegationGuard` SHALL provide a `RecordProviderFailure(provider string, success bool)` method that tracks provider-level failures using the same circuit breaker logic as agent delegation tracking. Provider keys SHALL be prefixed with `"provider:"` to avoid collision with agent names.

#### Scenario: Provider circuit opens after threshold failures
- **WHEN** a provider has consecutive failures exceeding the failure threshold
- **THEN** the guard SHALL mark that provider's circuit as open

#### Scenario: Provider circuit is independent of agent circuits
- **WHEN** provider "openai" circuit is open
- **THEN** agent "operator" circuit SHALL NOT be affected

### Requirement: Recovery decision event
The system SHALL publish a `RecoveryDecisionEvent` on the event bus when a recovery decision is made. The event SHALL include `CauseClass`, `Action`, `Attempt`, `Backoff` duration, and `SessionKey` fields.

#### Scenario: Recovery decision event published on retry
- **WHEN** recovery decides to retry
- **THEN** a `RecoveryDecisionEvent` SHALL be published with the cause class, action "retry", attempt number, computed backoff, and session ID

#### Scenario: Recovery decision event published on escalation
- **WHEN** recovery decides to escalate
- **THEN** a `RecoveryDecisionEvent` SHALL be published with action "escalate" and zero backoff
