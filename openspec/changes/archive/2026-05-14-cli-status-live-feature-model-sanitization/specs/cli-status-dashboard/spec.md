## MODIFIED Requirements

### Requirement: Unified status dashboard command
The system SHALL provide a `lango status` command that displays system health, configuration state, active channels, and feature status in a single dashboard view.

#### Scenario: Live server feature model text is replay-safe
- **WHEN** live `/health` feature names, reasons, or suggestions contain ANSI/OSC escape sequences or embedded newlines before status collection
- **THEN** the status command SHALL strip those control sequences
- **AND** it SHALL normalize the stored `serverInfo.features` text to a single line before replay or JSON output
