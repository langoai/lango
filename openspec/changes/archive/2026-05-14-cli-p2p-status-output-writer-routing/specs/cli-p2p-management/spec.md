## ADDED Requirements

### Requirement: P2P status output routing

`lango p2p status` SHALL write all non-error output through the Cobra command output stream so wrappers and test harnesses can capture text and JSON output without intercepting process-global stdout.

#### Scenario: Status text output writes to command output
- **WHEN** user runs `lango p2p status`
- **THEN** the command writes the status text output to the Cobra command output stream

#### Scenario: Status JSON output writes to command output
- **WHEN** user runs `lango p2p status --json`
- **THEN** the command writes the JSON payload to the Cobra command output stream
