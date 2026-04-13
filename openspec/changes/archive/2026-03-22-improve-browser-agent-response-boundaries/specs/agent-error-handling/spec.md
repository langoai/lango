## MODIFIED Requirements

### Requirement: User-facing error messages
The `AgentError` SHALL provide a `UserMessage()` method that returns a human-readable message including the error code and actionable guidance. User-facing messages SHALL NOT instruct the user to read a raw partial draft above.

#### Scenario: Timeout with partial result
- **WHEN** an `AgentError` has Code `ErrTimeout` and a non-empty `Partial` field
- **THEN** `UserMessage()` SHALL report the timeout with actionable guidance
- **AND** it SHALL NOT claim that the partial response was shown to the user

#### Scenario: Timeout without partial result
- **WHEN** an `AgentError` has Code `ErrTimeout` and an empty `Partial` field
- **THEN** `UserMessage()` SHALL suggest breaking the question into smaller parts

### Requirement: Partial result recovery in runAgent
When `runAgent()` receives an `AgentError` with a non-empty `Partial`, it SHALL retain the partial internally for diagnostics but SHALL NOT return the raw partial text to the user.

#### Scenario: Partial result suppressed from user response
- **WHEN** the agent returns an `AgentError` with `Partial` text
- **THEN** `runAgent()` SHALL return only a user-facing warning/error note
- **AND** it SHALL NOT append the raw partial draft to that message

#### Scenario: Error without partial propagated normally
- **WHEN** the agent returns an `AgentError` with empty `Partial`
- **THEN** `runAgent()` SHALL return the error to the channel for error display
