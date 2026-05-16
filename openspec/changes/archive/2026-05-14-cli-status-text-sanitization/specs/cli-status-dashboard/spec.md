## MODIFIED Requirements

### Requirement: Unified status dashboard command
The system SHALL provide a `lango status` command that displays system health, configuration state, active channels, and feature status in a single dashboard view.

#### Scenario: Rendered status CLI text stays plain and single-line
- **WHEN** provider/model labels, feature names/details, channel names, or dead-letter labels contain ANSI/OSC escape sequences or embedded newlines
- **THEN** the status CLI SHALL strip those control sequences
- **AND** it SHALL normalize the displayed text to a single line before rendering it
