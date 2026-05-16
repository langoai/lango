## MODIFIED Requirements

### Requirement: Unified status dashboard command
The system SHALL provide a `lango status` command that displays system health, configuration state, active channels, and feature status in a single dashboard view.

#### Scenario: Dead-letter summary model text is replay-safe
- **WHEN** dead-letter summary bucket labels, top latest reason labels, top latest actor labels, or top latest dispatch-reference labels contain ANSI/OSC escape sequences or embedded newlines before summary aggregation
- **THEN** the status command SHALL strip those control sequences
- **AND** it SHALL normalize the stored summary-model text to a single line before replay or JSON output
