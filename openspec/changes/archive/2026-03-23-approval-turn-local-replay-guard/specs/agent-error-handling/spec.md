## MODIFIED Requirements

### Requirement: Error classification
The system SHALL classify errors into codes: `ErrTimeout` (E001), `ErrModelError` (E002), `ErrToolError` (E003), `ErrTurnLimit` (E004), `ErrInternal` (E005). Classification SHALL be based on error content and context state.

#### Scenario: Context deadline classified as timeout
- **WHEN** the error is or wraps `context.DeadlineExceeded`
- **THEN** `classifyError` SHALL return `ErrTimeout`

#### Scenario: Turn limit error classified correctly
- **WHEN** the error message contains "maximum turn limit"
- **THEN** `classifyError` SHALL return `ErrTurnLimit`

#### Scenario: Approval failure classified as tool error
- **WHEN** the error wraps `approval.ErrDenied`, `approval.ErrTimeout`, or `approval.ErrUnavailable`
- **THEN** `classifyError` SHALL return `ErrToolError`

#### Scenario: Unknown error classified as internal
- **WHEN** the error does not match any known pattern
- **THEN** `classifyError` SHALL return `ErrInternal`

### Requirement: User-facing error messages
The `AgentError` SHALL provide a `UserMessage()` method that returns a human-readable message including the error code and actionable guidance. User-facing messages SHALL NOT instruct the user to read a raw partial draft above.

#### Scenario: Timeout with partial result
- **WHEN** an `AgentError` has Code `ErrTimeout` and a non-empty `Partial` field
- **THEN** `UserMessage()` SHALL report the timeout with actionable guidance
- **AND** it SHALL NOT claim that the partial response was shown to the user

#### Scenario: Approval denied message
- **WHEN** the underlying error wraps `approval.ErrDenied`
- **THEN** `UserMessage()` SHALL explain that the action was denied by approval

#### Scenario: Approval expired message
- **WHEN** the underlying error wraps `approval.ErrTimeout`
- **THEN** `UserMessage()` SHALL explain that the approval request expired

#### Scenario: Approval unavailable message
- **WHEN** the underlying error wraps `approval.ErrUnavailable`
- **THEN** `UserMessage()` SHALL explain that no approval channel was available
