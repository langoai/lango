## MODIFIED Requirements

### Requirement: Maximum turn limit per agent run
The system SHALL enforce a configurable maximum number of tool-calling turns per `Agent.Run()` invocation. The default limit SHALL be 50 turns. When the limit is reached, the system SHALL grant one wrap-up turn before yielding an error. Delegation events (TransferToAgent) SHALL NOT be counted as tool-calling turns.

#### Scenario: Turn limit reached with wrap-up
- **WHEN** the number of non-delegation function call events exceeds the configured maximum
- **THEN** the system SHALL log a warning, grant one wrap-up turn for the agent to finalize its response, and yield the current event
- **AND** if the agent exceeds the wrap-up turn, the system SHALL yield an error `"agent exceeded maximum turn limit (%d)"`

#### Scenario: Normal completion within limit
- **WHEN** the agent completes its work within the turn limit
- **THEN** all events SHALL be yielded normally with no interruption

#### Scenario: Custom turn limit via WithMaxTurns
- **WHEN** `WithMaxTurns(n)` is called with a positive value
- **THEN** the agent SHALL use `n` as the maximum turn limit instead of the default 50

#### Scenario: Zero or negative turn limit falls back to default
- **WHEN** `WithMaxTurns(0)` or `WithMaxTurns(-1)` is called
- **THEN** the agent SHALL use the default limit of 50
