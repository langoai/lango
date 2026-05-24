## MODIFIED Requirements

### Requirement: Unified status dashboard command
The system SHALL provide a `lango status` command that displays system health, configuration state, active channels, and feature status in a single dashboard view.

#### Scenario: JSON error payload stays plain and single-line
- **WHEN** `lango status` emits a JSON error payload and the underlying error text contains ANSI/OSC escape sequences or embedded newlines
- **THEN** the status command SHALL strip those control sequences
- **AND** it SHALL normalize the serialized `error` field to a single line before JSON output
