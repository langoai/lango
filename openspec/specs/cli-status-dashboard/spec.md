## Purpose

Unified status dashboard command (`lango status`) that combines health, configuration state, active channels, and feature status into a single view. Replaces the need to run health/doctor/metrics separately.

## Requirements

### Requirement: Unified status dashboard command
The system SHALL provide a `lango status` command that displays system health, configuration state, active channels, and feature status in a single dashboard view.

#### Scenario: Status with server not running
- **WHEN** user runs `lango status` and the server is not running
- **THEN** system displays config-based status (profile, gateway address, provider, features) with server marked as "not running"

#### Scenario: Status with server running
- **WHEN** user runs `lango status` and the server is running
- **THEN** system displays live health data alongside config-based status with server marked as "running"

#### Scenario: JSON output
- **WHEN** user runs `lango status --output json`
- **THEN** system outputs all status data as a JSON object with version, profile, serverUp, gateway, provider, model, features, channels, and serverInfo fields

### Requirement: Feature collection from config
The system SHALL collect feature status for 14 features: Knowledge, Embedding & RAG, Graph, Obs. Memory, Librarian, Multi-Agent, Cron, Background, Workflow, MCP, P2P, Payment, Economy, A2A.

#### Scenario: All features disabled
- **WHEN** default config is used
- **THEN** all optional features report as disabled

#### Scenario: MCP detail shows server count
- **WHEN** MCP is enabled with 2 servers configured
- **THEN** MCP feature detail shows "2 server(s)"

### Requirement: Channel collection
The system SHALL list active channels (telegram, discord, slack) based on their Enabled config flag.

#### Scenario: Multiple channels enabled
- **WHEN** Telegram and Slack are enabled in config
- **THEN** channels list contains "telegram" and "slack"
