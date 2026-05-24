## MODIFIED Requirements

### Requirement: Unified status dashboard command
The system SHALL provide a `lango status` command that displays system health, configuration state, active channels, and feature status in a single dashboard view.

#### Scenario: Root status table output uses the command writer
- **WHEN** `lango status` renders the human-readable dashboard
- **THEN** it SHALL write the dashboard through the Cobra command output writer
- **AND** wrappers or tests that replace `cmd.OutOrStdout()` SHALL capture the full dashboard output
