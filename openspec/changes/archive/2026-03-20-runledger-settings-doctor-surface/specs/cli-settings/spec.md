## MODIFIED Requirements

### Requirement: Configuration Coverage
The settings editor SHALL support editing all configuration sections, including RunLedger (Task OS) configuration.

#### Scenario: RunLedger category appears in Automation
- **WHEN** user launches `lango settings`
- **THEN** the `Automation` section SHALL include `RunLedger` alongside Cron Scheduler, Background Tasks, and Workflow Engine

## ADDED Requirements

### Requirement: RunLedger configuration form
The settings editor SHALL provide a RunLedger configuration form with the following fields:

- **Enabled** (`runledger_enabled`) — Boolean toggle
- **Shadow Mode** (`runledger_shadow`) — Boolean toggle
- **Write-Through** (`runledger_write_through`) — Boolean toggle
- **Authoritative Read** (`runledger_authoritative_read`) — Boolean toggle
- **Workspace Isolation** (`runledger_workspace_isolation`) — Boolean toggle
- **Stale TTL** (`runledger_stale_ttl`) — Duration text input
- **Max Run History** (`runledger_max_history`) — Integer input
- **Validator Timeout** (`runledger_validator_timeout`) — Duration text input
- **Planner Max Retries** (`runledger_planner_retries`) — Integer input

#### Scenario: Edit RunLedger settings
- **WHEN** user selects `RunLedger` from the settings menu
- **THEN** the editor SHALL display a form with all RunLedger fields pre-populated from `config.RunLedger`

#### Scenario: Save RunLedger settings
- **WHEN** user edits RunLedger fields and navigates back or saves
- **THEN** the config state SHALL be updated through `UpdateConfigFromForm`
- **AND** all edited values SHALL persist into `config.RunLedger`
