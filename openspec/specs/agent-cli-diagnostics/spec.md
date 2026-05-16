## Purpose

Capability spec for agent-cli-diagnostics. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: Trace list command
The system SHALL provide `lango agent trace list` that lists recent traces with outcomes. It SHALL support `--session`, `--limit`, `--outcome`, and `--output table|json` flags. It SHALL use bootstrap-aware pattern for DB access.

#### Scenario: List traces with outcome filter
- **WHEN** user runs `lango agent trace list --outcome timeout --limit 10`
- **THEN** the system SHALL display up to 10 traces with outcome `timeout`, showing trace ID, session key, outcome, duration, and timestamp

### Requirement: Trace detail command
The system SHALL provide `lango agent trace show <trace-id>` that displays a detailed event timeline for a specific trace, showing timestamp, event type, agent name, tool name, and payload excerpt.

#### Scenario: View trace timeline
- **WHEN** user runs `lango agent trace show abc-123`
- **THEN** the system SHALL display all events for trace `abc-123` ordered by sequence number

### Requirement: Delegation graph command
The system SHALL provide `lango agent graph <session-key>` that displays the delegation graph for a session, showing agents and handoff edges. It SHALL support `--output table|json`.

#### Scenario: View delegation graph
- **WHEN** user runs `lango agent graph tui-123456`
- **THEN** the system SHALL display all agents involved and delegation edges with counts

### Requirement: Trace metrics command
The system SHALL provide `lango agent trace metrics` that displays trace-derived per-agent performance metrics. It SHALL support `--output table|json` and `--agent` flags. This is distinct from the existing `lango metrics agents` command which shows token usage.

#### Scenario: View per-agent metrics
- **WHEN** user runs `lango agent trace metrics`
- **THEN** the system SHALL display per-agent success rate, turn count, p50/p95 duration

### Requirement: Trace metrics stays nested under the trace command tree
The system SHALL expose trace-derived metrics as `lango agent trace metrics` under the `trace` command tree rather than as a root-level `lango agent metrics` subcommand.

#### Scenario: Trace metrics remains nested under trace
- **WHEN** the operator inspects `lango agent --help` and `lango agent trace --help`
- **THEN** the metrics subcommand SHALL be reachable as `lango agent trace metrics`
- **AND** the root `lango agent` command SHALL NOT advertise a sibling `metrics` subcommand for trace-derived metrics

### Requirement: Bootstrap-aware agent command
`NewAgentCmd` SHALL accept both `cfgLoader` and `bootLoader` parameters. Commands that require DB access (trace, graph, trace metrics) SHALL use `bootLoader` for Ent client access, following the pattern used by `lango learning` and `lango security`.

#### Scenario: Commands work with DB
- **WHEN** user runs `lango agent trace list`
- **THEN** the system SHALL bootstrap the database connection and query the trace store

### Requirement: Agent diagnostics output format stays explicit and validated
`lango agent trace list`, `lango agent trace show`, `lango agent graph`, and `lango agent trace metrics` SHALL accept `--output table|json` and reject unknown values before bootstrap-dependent work.

#### Scenario: Agent diagnostics reject unknown output before bootstrap
- **WHEN** the operator runs one of those commands with `--output yaml`
- **THEN** the command SHALL return an actionable unknown-output-format error
- **AND** it SHALL NOT invoke bootstrap-dependent work
