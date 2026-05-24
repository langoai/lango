## MODIFIED Requirements

### Requirement: Unified status dashboard command
The system SHALL provide a `lango status` command that displays system health, configuration state, active channels, and feature status in a single dashboard view.

#### Scenario: Collected status model text is replay-safe
- **WHEN** provider/model labels, feature names/details, channel names, or live feature reasons contain ANSI/OSC escape sequences or embedded newlines before status collection or enrichment
- **THEN** the status command SHALL strip those control sequences
- **AND** it SHALL normalize the stored status model text to a single line before replay or JSON output
