## ADDED Requirements

### Requirement: Recovery action mapping
The system SHALL provide a `RecoveryAction` type and a `recoveryActionFor(FailureClassification) RecoveryAction` function. The mapping SHALL use the `CauseClass` field to determine whether a failure is retryable. Provider rate limit, transient, and connection causes SHALL map to `Retry`. Provider auth causes SHALL map to `AbortWithHint`. All other causes SHALL return nil (no recovery action).

#### Scenario: Rate limit maps to Retry
- **WHEN** `recoveryActionFor` is called with `CauseClass == "provider_rate_limit"`
- **THEN** it SHALL return `Retry`

#### Scenario: Connection error maps to Retry
- **WHEN** `recoveryActionFor` is called with `CauseClass == "provider_connection"`
- **THEN** it SHALL return `Retry`

#### Scenario: Auth error maps to AbortWithHint
- **WHEN** `recoveryActionFor` is called with `CauseClass == "provider_auth"`
- **THEN** it SHALL return `AbortWithHint`

#### Scenario: Approval denied maps to nil
- **WHEN** `recoveryActionFor` is called with `CauseClass == "approval_denied"`
- **THEN** it SHALL return nil
