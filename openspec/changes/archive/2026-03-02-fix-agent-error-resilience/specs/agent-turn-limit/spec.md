## MODIFIED Requirements

### Requirement: Maximum turn limit per agent run
The system SHALL enforce a configurable maximum number of tool-calling turns per `Agent.Run()` invocation. The default limit SHALL be 25 turns. When the limit is reached, the system SHALL grant one wrap-up turn before yielding an error. Delegation events (TransferToAgent) SHALL NOT be counted as tool-calling turns.

#### Scenario: Turn limit reached with wrap-up
- **WHEN** the number of non-delegation function call events exceeds the configured maximum
- **THEN** the system SHALL log a warning, grant one wrap-up turn for the agent to finalize its response, and yield the current event
- **AND** if the agent exceeds the wrap-up turn, the system SHALL yield an error `"agent exceeded maximum turn limit (%d)"`

#### Scenario: Normal completion within limit
- **WHEN** the agent completes its work within the turn limit
- **THEN** all events SHALL be yielded normally with no interruption

#### Scenario: Custom turn limit via WithMaxTurns
- **WHEN** `WithMaxTurns(n)` is called with a positive value
- **THEN** the agent SHALL use `n` as the maximum turn limit instead of the default 25

#### Scenario: Zero or negative turn limit falls back to default
- **WHEN** `WithMaxTurns(0)` or `WithMaxTurns(-1)` is called
- **THEN** the agent SHALL use the default limit of 25

### Requirement: Function call detection in events
The system SHALL count only events that contain at least one `FunctionCall` part as tool-calling turns.

#### Scenario: Event with function call parts
- **WHEN** an event's Content contains one or more parts with a non-nil `FunctionCall`
- **THEN** it SHALL be counted as a tool-calling turn

#### Scenario: Event without function calls
- **WHEN** an event contains only text parts or no parts
- **THEN** it SHALL NOT be counted as a tool-calling turn

## ADDED Requirements

### Requirement: Delegation event exclusion from turn counting
The system SHALL NOT count events that represent agent-to-agent delegation transfers as tool-calling turns. An event is a delegation event when its `Actions.TransferToAgent` field is non-empty.

#### Scenario: Delegation event not counted as turn
- **WHEN** an event contains FunctionCall parts AND has a non-empty `Actions.TransferToAgent`
- **THEN** it SHALL NOT be counted toward the turn limit

#### Scenario: Normal function call event counted
- **WHEN** an event contains FunctionCall parts AND has an empty `Actions.TransferToAgent`
- **THEN** it SHALL be counted toward the turn limit

### Requirement: Graceful wrap-up turn
The system SHALL grant exactly one wrap-up turn after the turn limit is reached, allowing the agent to finalize its response before hard stop.

#### Scenario: Wrap-up turn granted after limit reached
- **WHEN** the turn count exceeds maxTurns for the first time
- **THEN** the system SHALL log a warning with "granting wrap-up turn", yield the current event, and continue for one more iteration

#### Scenario: Hard stop after wrap-up turn consumed
- **WHEN** the turn count exceeds maxTurns and the wrap-up turn has already been granted
- **THEN** the system SHALL yield an error and stop iteration

### Requirement: Turn limit warning at 80% threshold
The system SHALL log a warning when the turn count reaches 80% of the configured maximum, providing observability into turn consumption.

#### Scenario: Warning logged at 80% of turn limit
- **WHEN** the turn count equals 80% of maxTurns (calculated as `maxTurns * 4 / 5`)
- **THEN** the system SHALL log a warning with session ID, current turn count, and max turns
- **AND** the warning SHALL be logged only once per agent run
