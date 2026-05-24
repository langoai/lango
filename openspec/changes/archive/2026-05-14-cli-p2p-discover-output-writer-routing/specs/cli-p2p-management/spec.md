## ADDED Requirements

### Requirement: P2P discover output routing

`lango p2p discover` SHALL write all non-error output through the Cobra command output stream so wrappers and test harnesses can capture empty-state, table, and JSON output without intercepting process-global stdout.

#### Scenario: Discover empty-state writes to command output
- **WHEN** user runs `lango p2p discover` and no agents are known
- **THEN** the command writes `No agents discovered. Try connecting to bootstrap peers first.` to the Cobra command output stream

#### Scenario: Discover table writes to command output
- **WHEN** user runs `lango p2p discover --tag research`
- **THEN** the command writes the discovery table to the Cobra command output stream

#### Scenario: Discover JSON writes to command output
- **WHEN** user runs `lango p2p discover --json`
- **THEN** the command writes the JSON payload to the Cobra command output stream
