# turn-retry-loop Specification

## Purpose
TBD - created by archiving change ux-elastic-turns. Update Purpose after archive.
## Requirements
### Requirement: RecoveryAction type and mapping
The system SHALL define a `RecoveryAction` type with values `Retry` and `AbortWithHint`. A `recoveryActionFor(FailureClassification) RecoveryAction` function SHALL map cause classes to actions: `CauseProviderRateLimit` → `Retry`, `CauseProviderTransient` → `Retry`, `CauseProviderConnection` → `Retry`, `CauseProviderAuth` → `AbortWithHint`. Unmapped causes SHALL return `nil` (no retry).

#### Scenario: Rate limit triggers retry
- **WHEN** `recoveryActionFor` is called with `CauseClass == CauseProviderRateLimit`
- **THEN** it SHALL return `Retry`

#### Scenario: Auth error triggers abort
- **WHEN** `recoveryActionFor` is called with `CauseClass == CauseProviderAuth`
- **THEN** it SHALL return `AbortWithHint`

#### Scenario: Unknown cause returns nil
- **WHEN** `recoveryActionFor` is called with an unmapped CauseClass (e.g., `CauseApprovalDenied`)
- **THEN** it SHALL return `nil`

### Requirement: Runner retry loop with per-attempt context
The `Runner.Run()` method SHALL wrap `executor.RunStreamingDetailed()` in a retry loop. Each attempt SHALL create a fresh `context.Context` derived from the parent. The loop SHALL execute at most `maxAttempts` (default 3) iterations. On `Retry`, the loop SHALL apply jittered exponential backoff before the next attempt. On `AbortWithHint` or `nil`, the loop SHALL exit immediately.

#### Scenario: Transient error retried with backoff
- **WHEN** the first attempt fails with `CauseProviderTransient` and `recoveryActionFor` returns `Retry`
- **THEN** the runner SHALL wait with jittered exponential backoff
- **AND** SHALL create a new context for the second attempt
- **AND** SHALL call `executor.RunStreamingDetailed()` again

#### Scenario: Max attempts exhausted
- **WHEN** all 3 attempts fail with retryable causes
- **THEN** the runner SHALL return the result from the final attempt

#### Scenario: Non-retryable error exits immediately
- **WHEN** the first attempt fails with `CauseProviderAuth`
- **THEN** the runner SHALL NOT retry and SHALL return the result immediately

#### Scenario: Successful attempt exits loop
- **WHEN** the first attempt succeeds
- **THEN** the runner SHALL return the successful result without further attempts

### Requirement: Recovery event emission from Runner
The Runner SHALL emit a `RecoveryInfo` event for each retry attempt. The event SHALL include the attempt number, cause class, and backoff duration. Events SHALL be recorded via `traceRecorder.recordRecovery()` within the same turn trace.

#### Scenario: Recovery event on retry
- **WHEN** the runner retries after a `CauseProviderRateLimit` failure
- **THEN** a `RecoveryInfo` event SHALL be emitted with the attempt number and cause class
- **AND** the event SHALL be recorded in the turn trace via `recordRecovery()`

