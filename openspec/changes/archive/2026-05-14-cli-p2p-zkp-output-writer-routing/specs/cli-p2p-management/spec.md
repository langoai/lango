## ADDED Requirements

### Requirement: P2P zkp output routing

`lango p2p zkp status` and `lango p2p zkp circuits` SHALL write all non-error output through the Cobra command output stream so wrappers and test harnesses can capture text, table, and JSON output without intercepting process-global stdout.

#### Scenario: ZKP status text output writes to command output
- **WHEN** user runs `lango p2p zkp status`
- **THEN** the command writes the status text output to the Cobra command output stream

#### Scenario: ZKP status JSON output writes to command output
- **WHEN** user runs `lango p2p zkp status --json`
- **THEN** the command writes the JSON payload to the Cobra command output stream

#### Scenario: ZKP circuits text output writes to command output
- **WHEN** user runs `lango p2p zkp circuits`
- **THEN** the command writes the circuits table to the Cobra command output stream

#### Scenario: ZKP circuits JSON output writes to command output
- **WHEN** user runs `lango p2p zkp circuits --json`
- **THEN** the command writes the JSON payload to the Cobra command output stream
