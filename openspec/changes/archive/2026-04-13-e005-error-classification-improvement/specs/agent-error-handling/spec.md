## ADDED Requirement: Provider auth error classification

#### Scenario: Provider auth error classified as model error

- **WHEN** the error message contains `"401"`, `"403"`, `"unauthorized"`, `"invalid api key"`, `"invalid_api_key"`, or `"authentication failed"` (case-insensitive)
- **THEN** `classifyError` SHALL return `ErrModelError` with `CauseClass` = `"provider_auth"`

#### Scenario: Case-insensitive auth matching

- **WHEN** the error message contains `"UNAUTHORIZED"` or `"Invalid API Key"` (mixed case)
- **THEN** `classifyError` SHALL still return `ErrModelError` with `CauseClass` = `"provider_auth"`

## ADDED Requirement: Provider connection error classification

#### Scenario: Provider connection error classified as model error

- **WHEN** the error message contains `"connection refused"`, `"no such host"`, `"dial tcp"`, or `"connection reset"` (case-insensitive)
- **THEN** `classifyError` SHALL return `ErrModelError` with `CauseClass` = `"provider_connection"`

## ADDED Requirement: Curated user messages for provider errors

User-facing messages for provider auth and connection errors SHALL provide specific actionable guidance without exposing raw error details.

#### Scenario: Auth error user message

- **WHEN** an `AgentError` has `CauseClass` = `"provider_auth"`
- **THEN** `UserMessage()` SHALL return a message about API key configuration
- **AND** SHALL NOT expose raw error details from `CauseDetail`

#### Scenario: Connection error user message

- **WHEN** an `AgentError` has `CauseClass` = `"provider_connection"`
- **THEN** `UserMessage()` SHALL return a message about network connectivity and provider URL
- **AND** SHALL NOT expose raw error details from `CauseDetail`

## ADDED Requirement: RecoveryPolicy for provider auth and connection errors

#### Scenario: Provider auth error escalates immediately

- **WHEN** inner executor returns `ErrModelError` with `CauseClass` = `"provider_auth"`
- **THEN** `RecoveryPolicy` SHALL return `RecoveryEscalate` (no retry)

#### Scenario: Provider connection error retries

- **WHEN** inner executor returns `ErrModelError` with `CauseClass` = `"provider_connection"` and recovery budget allows
- **THEN** `RecoveryPolicy` SHALL return `RecoveryRetry`

## MODIFIED Requirement: E005 fallback preserves operator diagnostics

_Modifies: "Unknown error classified as internal" scenario (main spec line 50-52)_

#### Scenario: Unknown error includes cause detail in operator summary

- **WHEN** the error does not match any known pattern
- **THEN** `classifyError` SHALL return `ErrInternal` with `CauseDetail` set to the error message
- **AND** `OperatorSummary` SHALL include a truncated version of the error message (max 200 chars)
- **AND** the user-facing `UserMessage()` SHALL NOT include the raw error message
