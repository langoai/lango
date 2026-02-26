## ADDED Requirements

### Requirement: Learning-based error correction on agent failure
The system SHALL support an optional `ErrorFixProvider` that returns known fixes for tool errors. When set and the initial agent run fails, the agent SHALL attempt one retry with the suggested fix.

#### Scenario: Error fix provider configured and fix available
- **WHEN** `WithErrorFixProvider` has been called with a non-nil provider
- **AND** the initial run fails with an error
- **AND** the provider returns a fix with `ok == true`
- **THEN** the agent SHALL retry with a correction message containing the original error and suggested fix

#### Scenario: Retry succeeds
- **WHEN** the retry with a learned fix succeeds
- **THEN** the agent SHALL return the retry response as the final result

#### Scenario: Retry fails
- **WHEN** the retry with a learned fix also fails
- **THEN** the agent SHALL log a warning and continue with the original error handling path

#### Scenario: No fix available
- **WHEN** the provider returns `ok == false` for the error
- **THEN** the agent SHALL proceed with normal error handling without retrying

#### Scenario: No error fix provider configured
- **WHEN** `WithErrorFixProvider` has not been called
- **THEN** the agent SHALL skip the self-correction path entirely

### Requirement: ErrorFixProvider interface
The `ErrorFixProvider` interface SHALL define `GetFixForError(ctx, toolName, err) (string, bool)` that returns a fix suggestion and whether one was found.

#### Scenario: Interface compliance with learning.Engine
- **WHEN** `learning.Engine` implements `GetFixForError`
- **THEN** it SHALL satisfy the `ErrorFixProvider` interface
