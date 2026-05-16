## ADDED Requirements

### Requirement: P2P workspace output routing

`lango p2p workspace` guidance commands SHALL write all non-error output through the Cobra command output stream so wrappers and test harnesses can capture text and JSON guidance output without intercepting process-global stdout.

#### Scenario: Workspace create text output writes to command output
- **WHEN** user runs `lango p2p workspace create <name>`
- **THEN** the command writes the guidance text to the Cobra command output stream

#### Scenario: Workspace create JSON output writes to command output
- **WHEN** user runs `lango p2p workspace create <name> --json`
- **THEN** the command writes the JSON payload to the Cobra command output stream

#### Scenario: Workspace list text output writes to command output
- **WHEN** user runs `lango p2p workspace list`
- **THEN** the command writes the guidance text to the Cobra command output stream

#### Scenario: Workspace list JSON output writes to command output
- **WHEN** user runs `lango p2p workspace list --json`
- **THEN** the command writes the JSON payload to the Cobra command output stream

#### Scenario: Workspace status text output writes to command output
- **WHEN** user runs `lango p2p workspace status <workspace-id>`
- **THEN** the command writes the guidance text to the Cobra command output stream

#### Scenario: Workspace status JSON output writes to command output
- **WHEN** user runs `lango p2p workspace status <workspace-id> --json`
- **THEN** the command writes the JSON payload to the Cobra command output stream

#### Scenario: Workspace join/leave text output writes to command output
- **WHEN** user runs `lango p2p workspace join <workspace-id>` or `leave <workspace-id>`
- **THEN** the command writes the guidance text to the Cobra command output stream
