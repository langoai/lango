## MODIFIED Requirements

### Requirement: Unified status dashboard command
The system SHALL provide a `lango status` command that displays system health, configuration state, active channels, and feature status in a single dashboard view.

#### Scenario: Non-JSON status command errors stay plain and single-line
- **WHEN** any downstream status command error contains ANSI/OSC escape sequences or embedded newlines before non-JSON CLI handling
- **THEN** the status command SHALL strip those control sequences at the shared non-JSON error boundary
- **AND** it SHALL normalize the surfaced error text to a single line
- **AND** it SHALL preserve the original error as the unwrap cause
