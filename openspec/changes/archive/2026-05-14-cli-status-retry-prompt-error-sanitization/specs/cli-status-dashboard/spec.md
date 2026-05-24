## MODIFIED Requirements

### Requirement: Unified status dashboard command
The system SHALL provide a `lango status` command that displays system health, configuration state, active channels, and feature status in a single dashboard view.

#### Scenario: Dead-letter retry prompt and non-JSON error text stay plain and single-line
- **WHEN** a transaction receipt ID contains ANSI/OSC escape sequences or embedded newlines before dead-letter retry confirmation or non-JSON error reporting
- **THEN** the status command SHALL strip those control sequences from the displayed transaction label
- **AND** it SHALL normalize the prompt and error text to a single line while preserving the lookup input passed to the retry flow
