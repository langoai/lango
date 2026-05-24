## ADDED Requirements

### Requirement: P2P peers output routing

`lango p2p peers` SHALL write all non-error output through the Cobra command output stream so wrappers and test harnesses can capture empty-state, table, and JSON output without intercepting process-global stdout.

#### Scenario: No connected peers writes to command output
- **WHEN** user runs `lango p2p peers` with no connected peers
- **THEN** the command writes `No connected peers.` to the Cobra command output stream

#### Scenario: Peers table writes to command output
- **WHEN** user runs `lango p2p peers` with connected peers
- **THEN** the command writes the peers table to the Cobra command output stream

#### Scenario: Peers JSON writes to command output
- **WHEN** user runs `lango p2p peers --json`
- **THEN** the command writes the JSON payload to the Cobra command output stream
