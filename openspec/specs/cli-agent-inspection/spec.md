## Purpose

Define the CLI commands for inspecting agent mode, configuration, and listing local/remote agents.
## Requirements
### Requirement: Agent status command

The existing `lango agent status` command contract remains preserved unless explicitly changed by this requirement: current status fields continue to be shown in table and JSON output, and this change only adds teammate runtime reporting. The command SHALL expose `teammate_runtime` when multi-agent mode is enabled. For the production in-process teammate runtime defined by this change, `dynamic-v1` means the built-in teammate runtime path is configured and available for built-in teammates under multi-agent mode; it does not imply that legacy static fallback or remote A2A paths are disabled. If that built-in runtime path is not configured or not available, the command SHALL omit the `teammate_runtime` field rather than reporting `dynamic-v1`.

#### Scenario: Table output shows dynamic teammate runtime
- **WHEN** `lango agent status` is run with `agent.multiAgent: true`
- **AND** the built-in dynamic teammate runtime path is configured and available
- **THEN** the command SHALL display a `Teammate Runtime` field with value `dynamic-v1`

#### Scenario: JSON output shows dynamic teammate runtime
- **WHEN** `lango agent status --json` is run with `agent.multiAgent: true`
- **AND** the built-in dynamic teammate runtime path is configured and available
- **THEN** the output SHALL include `"teammate_runtime": "dynamic-v1"`

#### Scenario: Single-agent output omits teammate runtime
- **WHEN** `lango agent status` is run with `agent.multiAgent: false`
- **THEN** the command SHALL omit the teammate runtime field

#### Scenario: Multi-agent without built-in dynamic runtime omits teammate runtime
- **WHEN** `lango agent status` is run with `agent.multiAgent: true`
- **AND** the built-in dynamic teammate runtime path is not configured or not available
- **THEN** the command SHALL omit the teammate runtime field

### Requirement: Performance fields in agent status
`lango agent status` SHALL display MaxTurns, ErrorCorrectionEnabled, and MaxDelegationRounds (multi-agent only) with their effective values (config or default).

#### Scenario: Default values displayed in single-agent mode
- **WHEN** user runs `lango agent status` with no performance config and `agent.multiAgent: false`
- **THEN** output SHALL show Max Turns: 50, Error Correction: true

#### Scenario: Default values displayed in multi-agent mode
- **WHEN** user runs `lango agent status` with no performance config and `agent.multiAgent: true`
- **THEN** output SHALL show Max Turns: 75, Error Correction: true
- **THEN** output SHALL include Delegation Rounds field

#### Scenario: JSON output includes new fields
- **WHEN** user runs `lango agent status --json`
- **THEN** JSON output SHALL include `max_turns`, `error_correction_enabled`, and `max_delegation_rounds` fields

### Requirement: Agent list displays registry sources
The `lango agent list` command SHALL load agents from the dynamic agent registry (embedded + user-defined stores) instead of hardcoded lists. Each agent entry SHALL display its source: "builtin", "embedded", "user", or "remote". The command SHALL support `--json` and `--check` flags.

#### Scenario: List shows embedded agents
- **WHEN** `lango agent list` is run with no user-defined agents
- **THEN** it SHALL display the 8 default agents with source "embedded"

#### Scenario: List shows user-defined agents
- **WHEN** user-defined agents exist in the configured agents directory
- **THEN** they SHALL appear in the list with source "user"

#### Scenario: List shows remote A2A agents
- **WHEN** A2A remote agents are configured
- **THEN** they SHALL appear in a separate table with source "a2a" and URL

#### Scenario: JSON output includes source
- **WHEN** `lango agent list --json` is run
- **THEN** each entry SHALL include "type" ("local" or "remote") and "source" fields

#### Scenario: Check connectivity
- **WHEN** user runs `lango agent list --check` with remote agents
- **THEN** system tests connectivity to each remote agent (2s timeout) and adds STATUS column showing "ok" or "unreachable"

### Requirement: Agent status shows registry info
The `lango agent status` command SHALL display registry information including builtin agent count, user agent count, active agent count, and agents directory path.

#### Scenario: Status includes registry counts
- **WHEN** `lango agent status` is run
- **THEN** it SHALL display "Builtin Agents", "User Agents", "Active Agents" counts

#### Scenario: Status shows P2P and hooks status
- **WHEN** `lango agent status` is run
- **THEN** it SHALL display P2P enabled status and Hooks enabled status

#### Scenario: JSON status includes registry
- **WHEN** `lango agent status --json` is run
- **THEN** the output SHALL include a "registry" object with builtin, user, active counts

