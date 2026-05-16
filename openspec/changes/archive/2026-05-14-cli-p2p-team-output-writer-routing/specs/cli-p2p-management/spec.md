## ADDED Requirements

### Requirement: P2P team output routing

`lango p2p team` guidance commands SHALL write all non-error output through the Cobra command output stream so wrappers and test harnesses can capture text and JSON guidance output without intercepting process-global stdout.

#### Scenario: Team list text output writes to command output
- **WHEN** user runs `lango p2p team list`
- **THEN** the command writes the guidance text to the Cobra command output stream

#### Scenario: Team list JSON output writes to command output
- **WHEN** user runs `lango p2p team list --json`
- **THEN** the command writes the JSON payload to the Cobra command output stream

#### Scenario: Team status text output writes to command output
- **WHEN** user runs `lango p2p team status <team-id>`
- **THEN** the command writes the guidance text to the Cobra command output stream

#### Scenario: Team status JSON output writes to command output
- **WHEN** user runs `lango p2p team status <team-id> --json`
- **THEN** the command writes the JSON payload to the Cobra command output stream

#### Scenario: Team disband text output writes to command output
- **WHEN** user runs `lango p2p team disband <team-id>`
- **THEN** the command writes the guidance text to the Cobra command output stream
