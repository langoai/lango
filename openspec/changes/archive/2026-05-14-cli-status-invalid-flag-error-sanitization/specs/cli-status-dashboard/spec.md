## MODIFIED Requirements

### Requirement: Unified status dashboard command
The system SHALL provide a `lango status` command that displays system health, configuration state, active channels, and feature status in a single dashboard view.

#### Scenario: Dead-letter invalid-flag validation errors stay plain and single-line
- **WHEN** invalid dead-letter subtype, family, or RFC3339 flag values contain ANSI/OSC escape sequences or embedded newlines before CLI validation fails
- **THEN** the status command SHALL strip those control sequences from the echoed invalid value
- **AND** it SHALL normalize the non-JSON validation error text to a single line
