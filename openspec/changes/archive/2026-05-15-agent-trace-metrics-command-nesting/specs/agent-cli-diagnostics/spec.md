## MODIFIED Requirements

### Requirement: Trace detail command
The system SHALL provide `lango agent trace show <trace-id>` that displays a detailed event timeline for a specific trace, showing timestamp, event type, agent name, tool name, and payload excerpt.

#### Scenario: View trace timeline
- **WHEN** user runs `lango agent trace show abc-123`
- **THEN** the system SHALL display all events for trace `abc-123` ordered by sequence number

### Requirement: Trace metrics stays nested under the trace command tree
The system SHALL expose trace-derived metrics as `lango agent trace metrics` under the `trace` command tree rather than as a root-level `lango agent metrics` subcommand.

#### Scenario: Trace metrics remains nested under trace
- **WHEN** the operator inspects `lango agent --help` and `lango agent trace --help`
- **THEN** the metrics subcommand SHALL be reachable as `lango agent trace metrics`
- **AND** the root `lango agent` command SHALL NOT advertise a sibling `metrics` subcommand for trace-derived metrics
