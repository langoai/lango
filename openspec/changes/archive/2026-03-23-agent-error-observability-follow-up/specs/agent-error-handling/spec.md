## ADDED Requirements

### Requirement: Failure classification preserves operator diagnostics
The system SHALL classify agent failures with structured operator-facing metadata in addition to broad user-facing error codes.

#### Scenario: Agent error carries cause metadata
- **WHEN** the runtime classifies a failure
- **THEN** the resulting `AgentError` SHALL include `CauseClass`, `CauseDetail`, and `OperatorSummary`
- **AND** the user-facing message MAY remain broader than the operator-facing summary

#### Scenario: Sentinel errors take precedence
- **WHEN** an error wraps a known sentinel such as approval denial or timeout
- **THEN** the classification SHALL use the sentinel-derived `CauseClass`
- **AND** SHALL NOT fall through to a generic heuristic cause

### Requirement: Turn-limit failures have distinct cause class
Turn-limit failures SHALL be classified distinctly from repeated-call failures.

#### Scenario: Turn limit maps to turn_limit_exceeded
- **WHEN** a run fails because it exceeded the configured maximum turn limit
- **THEN** the failure SHALL use `ErrTurnLimit`
- **AND** its `CauseClass` SHALL be `turn_limit_exceeded`
