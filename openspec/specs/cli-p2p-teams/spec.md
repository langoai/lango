# CLI P2P Teams

## Purpose
Provides CLI commands for managing P2P teams, including listing teams, viewing team status, and disbanding teams.

## Requirements

### Requirement: P2P team list command
The system SHALL provide a `lango p2p team list [--output table|json]` guidance command that describes how to inspect runtime-backed P2P teams. The command SHALL use bootLoader for config access but SHALL NOT initialize a full P2P node. When P2P is disabled, the command SHALL return a clear error message.

#### Scenario: List teams with JSON output
- **WHEN** user runs `lango p2p team list --output json`
- **THEN** system outputs an empty JSON array when no runtime team snapshot is available

#### Scenario: P2P disabled
- **WHEN** user runs `lango p2p team list` with `p2p.enabled` set to false
- **THEN** system returns error "P2P networking is not enabled (set p2p.enabled = true)"

### Requirement: P2P team status command
The system SHALL provide a `lango p2p team status <name> [--output table|json]` guidance command that explains how to inspect a runtime-backed P2P team.

#### Scenario: Team inspection guidance
- **WHEN** user runs `lango p2p team status my-team`
- **THEN** system explains that teams are runtime-only structures and points the operator to the concrete `team_status` tool

#### Scenario: Team status with JSON output
- **WHEN** user runs `lango p2p team status nonexistent --output json`
- **THEN** system outputs a JSON object containing `"error": "team not found (teams are runtime-only)"`

### Requirement: P2P team disband command
The system SHALL provide a `lango p2p team disband <name> [--force]` command that disbands a P2P team. The command SHALL prompt for confirmation unless `--force` is provided.

#### Scenario: Disband with confirmation
- **WHEN** user runs `lango p2p team disband my-team` and confirms with "y"
- **THEN** system disbands the team and prints "Team 'my-team' disbanded"

#### Scenario: Force disband
- **WHEN** user runs `lango p2p team disband my-team --force`
- **THEN** system disbands the team without prompting

### Requirement: P2P team command group entry
The system SHALL provide a `lango p2p team` command group that shows help text listing all team subcommands when invoked without a subcommand.

#### Scenario: Help text
- **WHEN** user runs `lango p2p team`
- **THEN** system displays help listing list, status, and disband subcommands
