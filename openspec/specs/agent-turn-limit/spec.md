## ADDED Requirements

### Requirement: Maximum turn limit per agent run
The system SHALL enforce a configurable maximum number of tool-calling turns per `Agent.Run()` invocation. The default limit SHALL be 25 turns.

#### Scenario: Turn limit reached
- **WHEN** the number of events containing function calls exceeds the configured maximum
- **THEN** the system SHALL stop iterating, log a warning with session ID and turn counts, and yield an error `"agent exceeded maximum turn limit (%d)"`

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
