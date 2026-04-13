## ADDED Requirements

### Requirement: Extracted tool builders have parity coverage
The test suite SHALL include parity coverage for extracted tool builders so refactors do not silently change tool names or remove handlers.

#### Scenario: Extracted builders expose expected tool names
- **WHEN** the builder parity tests run
- **THEN** extracted builder functions SHALL return the expected tool names in stable order for the covered packages

#### Scenario: Extracted builders have non-nil handlers
- **WHEN** parity tests inspect tools returned by extracted builders
- **THEN** every tool SHALL have a non-nil handler

#### Scenario: Extracted builders avoid duplicate names
- **WHEN** parity tests inspect a builder result set
- **THEN** tool names within that result SHALL be unique
