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

### Requirement: REJECT text detection safety net
`RunAndCollect` SHALL detect `[REJECT]` text patterns in successful agent responses. When detected on an agent with sub-agents, it SHALL retry once with a system correction message instructing the orchestrator to re-evaluate and route to a different agent or answer directly.

#### Scenario: REJECT text detected in response
- **WHEN** `RunAndCollect` receives a successful response containing `[REJECT]`
- **AND** the agent has sub-agents (is an orchestrator)
- **THEN** it SHALL log a warning and retry with a correction message containing the original user input

#### Scenario: Retry succeeds without REJECT
- **WHEN** the retry produces a response without `[REJECT]` text
- **THEN** `RunAndCollect` SHALL return the retry response

#### Scenario: Retry also contains REJECT
- **WHEN** the retry response also contains `[REJECT]` text
- **THEN** `RunAndCollect` SHALL fall through and return the original response

#### Scenario: No sub-agents (single-agent mode)
- **WHEN** the agent has no sub-agents
- **AND** the response contains `[REJECT]` text
- **THEN** `RunAndCollect` SHALL NOT attempt a retry (safety net only applies to orchestrator)

#### Scenario: Normal response without REJECT
- **WHEN** the response does not contain `[REJECT]` text
- **THEN** `RunAndCollect` SHALL return the response immediately without retry

### Requirement: REJECT pattern matching
The system SHALL provide a `containsRejectPattern` function that matches the exact `[REJECT]` text marker using `strings.Contains`. The match SHALL be case-sensitive (lowercase `[reject]` SHALL NOT match).

#### Scenario: Exact REJECT marker matched
- **WHEN** text contains `[REJECT]`
- **THEN** `containsRejectPattern` SHALL return true

#### Scenario: Case-sensitive matching
- **WHEN** text contains `[reject]` (lowercase)
- **THEN** `containsRejectPattern` SHALL return false

#### Scenario: Normal text not matched
- **WHEN** text contains no `[REJECT]` marker
- **THEN** `containsRejectPattern` SHALL return false

### Requirement: ErrorFixProvider interface
The `ErrorFixProvider` interface SHALL define `GetFixForError(ctx, toolName, err) (string, bool)` that returns a fix suggestion and whether one was found.

#### Scenario: Interface compliance with learning.Engine
- **WHEN** `learning.Engine` implements `GetFixForError`
- **THEN** it SHALL satisfy the `ErrorFixProvider` interface
