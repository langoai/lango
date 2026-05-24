## MODIFIED Requirements

### Requirement: Unified status dashboard command
The system SHALL provide a `lango status` command that displays system health, configuration state, active channels, and feature status in a single dashboard view.

#### Scenario: Dead-letter list/detail JSON payload text is replay-safe
- **WHEN** dead-letter backlog JSON fields or dead-letter detail JSON fields contain ANSI/OSC escape sequences or embedded newlines before CLI serialization
- **THEN** the status command SHALL strip those control sequences
- **AND** it SHALL normalize the serialized JSON payload text to a single line while preserving the existing data shape
