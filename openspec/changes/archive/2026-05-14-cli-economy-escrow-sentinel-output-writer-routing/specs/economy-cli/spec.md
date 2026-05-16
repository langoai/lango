## ADDED Requirements

### Requirement: Escrow sentinel CLI output routing

`lango economy escrow sentinel status` SHALL write all non-error output through the Cobra command output stream so wrappers and test harnesses can capture active and disabled-state output without intercepting process-global stdout.

#### Scenario: Escrow sentinel active output writes to command output
- **WHEN** `lango economy escrow sentinel status` is run with on-chain escrow enabled
- **THEN** the command writes the sentinel status and operator guidance to the Cobra command output stream

#### Scenario: Escrow sentinel on-chain-disabled output writes to command output
- **WHEN** `lango economy escrow sentinel status` is run with on-chain escrow disabled
- **THEN** the command writes the on-chain-disabled message to the Cobra command output stream

#### Scenario: Escrow sentinel escrow-disabled output writes to command output
- **WHEN** `lango economy escrow sentinel status` is run with escrow disabled
- **THEN** the command writes the escrow-disabled message to the Cobra command output stream
