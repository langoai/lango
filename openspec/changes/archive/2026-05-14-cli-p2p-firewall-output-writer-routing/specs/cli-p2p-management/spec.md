## ADDED Requirements

### Requirement: P2P firewall output routing

`lango p2p firewall` subcommands SHALL write all non-error output through the Cobra command output stream so wrappers and test harnesses can capture empty-state, table, JSON, and guidance output without intercepting process-global stdout.

#### Scenario: Firewall list empty-state writes to command output
- **WHEN** user runs `lango p2p firewall list` with no configured rules
- **THEN** the command writes the empty-state message to the Cobra command output stream

#### Scenario: Firewall list table writes to command output
- **WHEN** user runs `lango p2p firewall list` with configured rules
- **THEN** the command writes the rules table to the Cobra command output stream

#### Scenario: Firewall list JSON writes to command output
- **WHEN** user runs `lango p2p firewall list --json`
- **THEN** the command writes the JSON payload to the Cobra command output stream

#### Scenario: Firewall add guidance writes to command output
- **WHEN** user runs `lango p2p firewall add ...`
- **THEN** the command writes the rule summary and persistence guidance to the Cobra command output stream

#### Scenario: Firewall remove guidance writes to command output
- **WHEN** user runs `lango p2p firewall remove <peer-did>`
- **THEN** the command writes the runtime removal guidance to the Cobra command output stream
