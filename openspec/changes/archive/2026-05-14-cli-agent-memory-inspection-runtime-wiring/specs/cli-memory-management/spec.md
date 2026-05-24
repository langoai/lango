## MODIFIED Requirements

### Requirement: Agent memory agents command
The system SHALL provide a `lango memory agents [--json]` command that lists agents with persistent memory entries from the configured session database.

#### Scenario: Agent memory agents lists persisted agent names
- **WHEN** user runs `lango memory agents`
- **THEN** the command outputs the persisted agent names with entry counts and last-updated timestamps

#### Scenario: Agent memory agents JSON output
- **WHEN** user runs `lango memory agents --json`
- **THEN** the command outputs a JSON array of agent summary objects

### Requirement: Agent memory single-agent command
The system SHALL provide a `lango memory agent <name> [--limit N] [--json]` command that lists stored memory entries for the specified agent from the configured session database.

#### Scenario: Agent memory command lists persisted entries
- **WHEN** user runs `lango memory agent <name>`
- **THEN** the command outputs that agent's stored memory entries in a table

#### Scenario: Agent memory command JSON output
- **WHEN** user runs `lango memory agent <name> --json`
- **THEN** the command outputs a JSON array of stored memory entries
