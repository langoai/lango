## MODIFIED Requirements

### Requirement: Error classification
The system SHALL classify errors into codes: `ErrTimeout` (E001), `ErrModelError` (E002), `ErrToolError` (E003), `ErrTurnLimit` (E004), `ErrInternal` (E005), `ErrIdleTimeout` (E006). Classification SHALL be based on error content and context state. Errors containing "thought_signature" or "thoughtSignature" in their message SHALL be classified as `ErrModelError` to prevent learning-based retry attempts.

#### Scenario: Context deadline classified as timeout
- **WHEN** the error is or wraps `context.DeadlineExceeded`
- **THEN** `classifyError` SHALL return `ErrTimeout`

#### Scenario: Turn limit error classified correctly
- **WHEN** the error message contains "maximum turn limit"
- **THEN** `classifyError` SHALL return `ErrTurnLimit`

#### Scenario: thought_signature error classified as model error
- **WHEN** the error message contains "thought_signature"
- **THEN** `classifyError` SHALL return `ErrModelError`

#### Scenario: thoughtSignature camelCase error classified as model error
- **WHEN** the error message contains "thoughtSignature"
- **THEN** `classifyError` SHALL return `ErrModelError`

#### Scenario: Unknown error classified as internal
- **WHEN** the error does not match any known pattern
- **THEN** `classifyError` SHALL return `ErrInternal`
