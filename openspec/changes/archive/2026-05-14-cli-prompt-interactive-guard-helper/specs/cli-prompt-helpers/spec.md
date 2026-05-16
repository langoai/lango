## ADDED Requirements

### Requirement: Shared interactive-terminal guard supports command-specific guidance
The shared CLI prompt package SHALL provide a helper that fails when the current stdin is not an interactive terminal and returns a caller-supplied guidance error.

#### Scenario: Interactive-terminal guard fails in non-interactive mode
- **WHEN** the shared interactive-terminal guard runs while stdin is not an interactive terminal
- **THEN** it SHALL return the caller-supplied error

#### Scenario: Interactive-terminal guard passes in interactive mode
- **WHEN** the shared interactive-terminal guard runs while stdin is interactive
- **THEN** it SHALL return nil
