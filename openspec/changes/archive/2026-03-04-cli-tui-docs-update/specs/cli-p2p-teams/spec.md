## ADDED Requirements

### Requirement: P2P team list command
The system SHALL provide a `lango p2p team list [--json]` command that displays all known P2P teams. The command SHALL use bootLoader for config access but SHALL NOT initialize a full P2P node. When P2P is disabled, the command SHALL return a clear error message.

#### Scenario: List teams with JSON output
- **WHEN** user runs `lango p2p team list --json`
- **THEN** system outputs a JSON array of team objects with fields: name, members, createdAt

#### Scenario: P2P disabled
- **WHEN** user runs `lango p2p team list` with `p2p.enabled` set to false
- **THEN** system returns error "P2P networking is not enabled (set p2p.enabled = true)"

### Requirement: P2P team status command
The system SHALL provide a `lango p2p team status <name> [--json]` command that displays detailed status for a specific P2P team, including member count, active connections, and team role assignments.

#### Scenario: Team exists
- **WHEN** user runs `lango p2p team status my-team`
- **THEN** system displays team name, member list with peer IDs, and connection status

#### Scenario: Team not found
- **WHEN** user runs `lango p2p team status nonexistent`
- **THEN** system returns error indicating the team was not found

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
