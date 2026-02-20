## MODIFIED Requirements

### Requirement: Tool Result Observation
The system SHALL observe every tool execution result to detect error patterns and track successes.

#### Scenario: Tool execution error
- **WHEN** `OnToolResult` is called with a non-nil error
- **THEN** the system SHALL extract the error pattern, categorize it, and store a learning entry

#### Scenario: Tool execution success
- **WHEN** `OnToolResult` is called with a nil error
- **THEN** the system SHALL search for related learnings by tool name and boost their confidence

#### Scenario: Skip duplicate high-confidence learnings
- **WHEN** an error occurs and a matching learning with confidence > 0.5 already exists
- **THEN** the system SHALL skip creating a new learning entry

#### Scenario: Error save failure logging
- **WHEN** saving a learning entry fails in `handleError`
- **THEN** the system SHALL log at Warn level with session key, tool name, and error details
