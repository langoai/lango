## Purpose

Define the CLI commands for inspecting agent mode, configuration, and listing local/remote agents.

## Requirements

### Requirement: Agent status command
The system SHALL provide a `lango agent status` command that displays agent mode (single/multi-agent), provider, model, and A2A configuration. The command SHALL support a `--json` flag.

#### Scenario: Single agent mode
- **WHEN** user runs `lango agent status` with multiAgent=false
- **THEN** system displays mode as "single" with provider and model info

#### Scenario: Multi-agent with A2A
- **WHEN** user runs `lango agent status` with multiAgent=true and A2A enabled
- **THEN** system displays mode as "multi-agent" with A2A base URL and agent name

### Requirement: Performance fields in agent status
`lango agent status` SHALL display MaxTurns, ErrorCorrectionEnabled, and MaxDelegationRounds (multi-agent only) with their effective values (config or default).

#### Scenario: Default values displayed
- **WHEN** user runs `lango agent status` with no performance config
- **THEN** output SHALL show Max Turns: 25, Error Correction: true

#### Scenario: Multi-agent delegation rounds
- **WHEN** user runs `lango agent status` with `agent.multiAgent: true`
- **THEN** output SHALL include Delegation Rounds field

#### Scenario: JSON output includes new fields
- **WHEN** user runs `lango agent status --json`
- **THEN** JSON output SHALL include `max_turns`, `error_correction_enabled`, and `max_delegation_rounds` fields

### Requirement: Agent list command
The system SHALL provide a `lango agent list` command that lists all local sub-agents and remote A2A agents. The command SHALL support `--json` and `--check` flags.

#### Scenario: List local agents
- **WHEN** user runs `lango agent list`
- **THEN** system displays NAME/TYPE/DESCRIPTION table for executor, researcher, planner, memory-manager

#### Scenario: List with remote agents
- **WHEN** user runs `lango agent list` with remote A2A agents configured
- **THEN** system displays local agents table and a separate remote agents table with NAME/TYPE/URL

#### Scenario: Check connectivity
- **WHEN** user runs `lango agent list --check` with remote agents
- **THEN** system tests connectivity to each remote agent (2s timeout) and adds STATUS column showing "ok" or "unreachable"
