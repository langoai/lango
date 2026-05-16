## ADDED Requirements

### Requirement: P2P identity output routing

`lango p2p identity` SHALL write all non-error output through the Cobra command output stream so wrappers and test harnesses can capture text and JSON output without intercepting process-global stdout.

#### Scenario: Identity text output writes to command output
- **WHEN** user runs `lango p2p identity`
- **THEN** the command writes the identity text output to the Cobra command output stream

#### Scenario: Identity JSON output writes to command output
- **WHEN** user runs `lango p2p identity --json`
- **THEN** the command writes the JSON payload to the Cobra command output stream
