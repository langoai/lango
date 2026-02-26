## MODIFIED Requirements

### Requirement: Tool Result Observation
The system SHALL observe every tool execution result to detect error patterns and track successes.

#### Scenario: Tool execution success — scoped confidence boost
- **WHEN** `OnToolResult` is called with a nil error
- **THEN** the system SHALL search for related learnings using the trigger `"tool:<toolName>"` and boost confidence ONLY for learnings whose trigger exactly matches

#### Scenario: Skip duplicate high-confidence learnings
- **WHEN** an error occurs and a matching learning with confidence > 0.7 already exists
- **THEN** the system SHALL skip creating a new learning entry

### Requirement: Auto-apply confidence threshold
The system SHALL use a confidence threshold of 0.7 (previously 0.5) for both `GetFixForError` and `handleError` skip-duplicate logic.

#### Scenario: GetFixForError returns fix above threshold
- **WHEN** a learning entity exists with confidence > 0.7 and a non-empty fix
- **THEN** `GetFixForError` SHALL return the fix with `ok == true`

#### Scenario: GetFixForError ignores low-confidence fix
- **WHEN** a learning entity exists with confidence <= 0.7
- **THEN** `GetFixForError` SHALL return `ok == false`

#### Scenario: Error handling skips known high-confidence learnings
- **WHEN** an error occurs and a matching learning has confidence > 0.7
- **THEN** `handleError` SHALL log the known fix and skip creating a new entry
