## ADDED Requirements

### Requirement: BudgetPolicy serialization
BudgetPolicy SHALL provide a `Serialize()` method that returns budget counters as `map[string]string` with keys `usage:budget_turns` and `usage:budget_delegations`, values as decimal string integers.

#### Scenario: Serialize captures current counters
- **WHEN** BudgetPolicy has recorded 5 turns and 3 delegations
- **THEN** `Serialize()` returns `{"usage:budget_turns": "5", "usage:budget_delegations": "3"}`

### Requirement: BudgetPolicy restoration
BudgetPolicy SHALL provide a `Restore(state map[string]string)` method that sets turn and delegation counters from the provided map. Missing or malformed keys SHALL be silently ignored (counters remain at their current value).

#### Scenario: Restore from valid state
- **WHEN** `Restore` is called with `{"usage:budget_turns": "10", "usage:budget_delegations": "4"}`
- **THEN** `TurnCount()` returns 10 and `DelegationCount()` returns 4

#### Scenario: Restore with missing keys
- **WHEN** `Restore` is called with an empty map
- **THEN** counters remain unchanged

#### Scenario: Restore with malformed values
- **WHEN** `Restore` is called with `{"usage:budget_turns": "abc"}`
- **THEN** turn counter remains unchanged (no error, no panic)

### Requirement: Lazy budget restoration on session resume
The system SHALL restore budget state from Session.Metadata on the first executor call for each session. Restoration SHALL happen at most once per session key (idempotent).

#### Scenario: First call restores budget
- **WHEN** a resumed session's first `RunStreamingDetailed` call executes
- **THEN** the executor reads Session.Metadata and calls `BudgetPolicy.Restore` before delegating to the inner executor

#### Scenario: Subsequent calls skip restoration
- **WHEN** a session has already been restored
- **THEN** subsequent `RunStreamingDetailed` calls proceed without re-reading Session.Metadata

### Requirement: Budget and token persistence after each turn
The system SHALL persist budget counters and cumulative token usage into Session.Metadata after each completed turn via an OnTurnComplete callback.

#### Scenario: Turn completion persists budget state
- **WHEN** a turn completes for a session
- **THEN** Session.Metadata contains `usage:budget_turns`, `usage:budget_delegations`, `usage:cumulative_input_tokens`, and `usage:cumulative_output_tokens` with current values

#### Scenario: Token values include collector metrics
- **WHEN** MetricsCollector has recorded 1000 input tokens and 500 output tokens for the session
- **THEN** `usage:cumulative_input_tokens` is at least "1000" and `usage:cumulative_output_tokens` is at least "500"

### Requirement: No schema changes to session storage
Budget persistence SHALL use the existing `Session.Metadata` field (map[string]string). No changes to session Store interface, EntStore, or ent schema are permitted.

#### Scenario: Metadata storage only
- **WHEN** budget persistence is active
- **THEN** only `Session.Metadata` keys prefixed with `usage:` are written; no new database columns or tables are created
