## ADDED Requirements

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
The system SHALL provide a `containsRejectPattern` function that matches the exact `[REJECT]` text marker using regex. The match SHALL be case-sensitive (lowercase `[reject]` SHALL NOT match).

#### Scenario: Exact REJECT marker matched
- **WHEN** text contains `[REJECT]`
- **THEN** `containsRejectPattern` SHALL return true

#### Scenario: Case-sensitive matching
- **WHEN** text contains `[reject]` (lowercase)
- **THEN** `containsRejectPattern` SHALL return false

#### Scenario: Normal text not matched
- **WHEN** text contains no `[REJECT]` marker
- **THEN** `containsRejectPattern` SHALL return false
