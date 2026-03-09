## ADDED Requirements

### Requirement: Structured agent error type
The system SHALL provide an `AgentError` type with fields: `Code` (ErrorCode), `Message` (string), `Cause` (error), `Partial` (string), and `Elapsed` (time.Duration). It SHALL implement the `error` and `Unwrap` interfaces.

#### Scenario: AgentError implements error interface
- **WHEN** an `AgentError` is created with Code `ErrTimeout` and Cause `context.DeadlineExceeded`
- **THEN** calling `Error()` SHALL return a string containing the error code and cause message

#### Scenario: AgentError supports errors.As unwrapping
- **WHEN** an `AgentError` is wrapped in `fmt.Errorf("outer: %w", agentErr)`
- **THEN** `errors.As(wrappedErr, &target)` SHALL succeed and populate the target with the original AgentError

### Requirement: Error classification
The system SHALL classify errors into codes: `ErrTimeout` (E001), `ErrModelError` (E002), `ErrToolError` (E003), `ErrTurnLimit` (E004), `ErrInternal` (E005). Classification SHALL be based on error content and context state.

#### Scenario: Context deadline classified as timeout
- **WHEN** the error is or wraps `context.DeadlineExceeded`
- **THEN** `classifyError` SHALL return `ErrTimeout`

#### Scenario: Turn limit error classified correctly
- **WHEN** the error message contains "maximum turn limit"
- **THEN** `classifyError` SHALL return `ErrTurnLimit`

#### Scenario: Unknown error classified as internal
- **WHEN** the error does not match any known pattern
- **THEN** `classifyError` SHALL return `ErrInternal`

### Requirement: User-facing error messages
The `AgentError` SHALL provide a `UserMessage()` method that returns a human-readable message including the error code and actionable guidance.

#### Scenario: Timeout with partial result
- **WHEN** an `AgentError` has Code `ErrTimeout` and a non-empty `Partial` field
- **THEN** `UserMessage()` SHALL mention that a partial response was recovered

#### Scenario: Timeout without partial result
- **WHEN** an `AgentError` has Code `ErrTimeout` and an empty `Partial` field
- **THEN** `UserMessage()` SHALL suggest breaking the question into smaller parts

### Requirement: Partial result preservation on agent error
When an agent run fails (timeout, turn limit, or other error), the system SHALL return the accumulated text as the `Partial` field of the `AgentError` instead of discarding it.

#### Scenario: Timeout preserves partial text
- **WHEN** the agent has accumulated text "Here is a partial..." and the context deadline fires
- **THEN** the returned `AgentError` SHALL have `Partial` equal to "Here is a partial..."

#### Scenario: Iterator error preserves partial text
- **WHEN** the agent iterator yields an error after producing some text chunks
- **THEN** the returned `AgentError` SHALL have `Partial` containing the accumulated chunks

### Requirement: Partial result recovery in runAgent
When `runAgent()` receives an `AgentError` with a non-empty `Partial`, it SHALL return the partial text appended with an error note as a successful response rather than propagating the error.

#### Scenario: Partial result returned as success
- **WHEN** the agent returns an `AgentError` with `Partial` "Here is my analysis..."
- **THEN** `runAgent()` SHALL return a string containing the partial text plus a warning note, and `nil` error

#### Scenario: Error without partial propagated normally
- **WHEN** the agent returns an `AgentError` with empty `Partial`
- **THEN** `runAgent()` SHALL return the error to the channel for error display

### Requirement: Channel error formatting
All channel `sendError()` functions SHALL use `formatChannelError()` which checks for a `UserMessage()` method via duck-typed interface assertion, falling back to `Error()` for plain errors.

#### Scenario: AgentError formatted with UserMessage
- **WHEN** a channel receives an error implementing `UserMessage()`
- **THEN** the displayed error SHALL use the `UserMessage()` output

#### Scenario: Plain error formatted with Error
- **WHEN** a channel receives a plain error without `UserMessage()`
- **THEN** the displayed error SHALL use `Error()` output prefixed with "Error:"
