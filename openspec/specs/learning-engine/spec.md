## ADDED Requirements

### Requirement: Tool Result Observation
The system SHALL observe every tool execution result to detect error patterns and track successes.

#### Scenario: Tool execution error
- **WHEN** `OnToolResult` is called with a non-nil error
- **THEN** the system SHALL extract the error pattern, categorize it, and store a learning entry

#### Scenario: Tool execution success — scoped confidence boost
- **WHEN** `OnToolResult` is called with a nil error
- **THEN** the system SHALL search for related learnings using the trigger `"tool:<toolName>"` and boost confidence ONLY for learnings whose trigger exactly matches

#### Scenario: Skip duplicate high-confidence learnings
- **WHEN** an error occurs and a matching learning with confidence > 0.7 already exists
- **THEN** the system SHALL skip creating a new learning entry

#### Scenario: Error save failure logging
- **WHEN** saving a learning entry fails in `handleError`
- **THEN** the system SHALL log at Warn level with session key, tool name, and error details

### Requirement: Error Pattern Extraction
The system SHALL normalize error messages into reusable patterns.

#### Scenario: Strip dynamic identifiers
- **WHEN** extracting an error pattern
- **THEN** the system SHALL replace UUIDs, timestamps, file paths, and port numbers with placeholders

### Requirement: Error Categorization
The system SHALL categorize errors into predefined categories.

#### Scenario: Timeout errors
- **WHEN** the error is a context deadline exceeded or contains "timeout"
- **THEN** the system SHALL categorize it as "timeout"
- **AND** SHALL use `errors.Is` for proper wrapped error detection of `context.DeadlineExceeded`

#### Scenario: Permission errors
- **WHEN** the error contains "permission denied", "access denied", or "forbidden"
- **THEN** the system SHALL categorize it as "permission"

#### Scenario: Provider errors
- **WHEN** the error contains "api", "model", "provider", or "rate limit"
- **THEN** the system SHALL categorize it as "provider_error"

#### Scenario: Tool errors
- **WHEN** a tool name is provided and the error does not match other categories
- **THEN** the system SHALL categorize it as "tool_error"

### Requirement: Parameter Summarization
The system SHALL summarize tool parameters before storing them in learnings.

#### Scenario: Truncate long strings
- **WHEN** a parameter value is a string longer than 200 characters
- **THEN** the system SHALL truncate it to 200 characters with "..." suffix

#### Scenario: Summarize arrays
- **WHEN** a parameter value is an array
- **THEN** the system SHALL replace it with a count summary (e.g., "[5 items]")

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

### Requirement: Confidence propagation uses float64 math
The system SHALL apply fractional confidence boosts when propagating success across similar learnings. BoostLearningConfidence SHALL accept a `confidenceBoost float64` parameter; when > 0, it adds the value directly to confidence and clamps to [0.1, 1.0]. When 0, existing success/occurrence ratio calculation is used.

#### Scenario: Graph engine propagates fractional confidence
- **WHEN** a tool succeeds and similar learnings exist in the graph
- **THEN** each similar learning's confidence SHALL increase by `0.1 * propagationRate` (0.03 for rate 0.3)

#### Scenario: Base engine uses existing ratio calculation
- **WHEN** the base engine boosts confidence on tool success
- **THEN** it SHALL call BoostLearningConfidence with confidenceBoost=0.0, using success/occurrence ratio

#### Scenario: Confidence clamps to valid range
- **WHEN** a confidence boost would result in a value outside [0.1, 1.0]
- **THEN** the value SHALL be clamped to [0.1, 1.0]
