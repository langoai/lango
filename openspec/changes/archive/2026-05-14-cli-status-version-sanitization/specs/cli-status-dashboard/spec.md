## MODIFIED Requirements

### Requirement: Unified status dashboard command
The system SHALL provide a `lango status` command that displays system health, configuration state, active channels, and feature status in a single dashboard view.

#### Scenario: Root status version text stays plain and single-line
- **WHEN** injected build version text contains ANSI/OSC escape sequences or embedded newlines before root status serialization
- **THEN** the status command SHALL strip those control sequences
- **AND** it SHALL normalize the stored top-level `version` field to a single line before table or JSON output
