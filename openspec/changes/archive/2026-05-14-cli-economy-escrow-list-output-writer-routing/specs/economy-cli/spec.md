## ADDED Requirements

### Requirement: Escrow CLI list output routing

`lango economy escrow list` SHALL write all non-error output through the Cobra command output stream so wrappers and test harnesses can capture summary and disabled-state output without intercepting process-global stdout.

#### Scenario: Escrow list summary writes to command output
- **WHEN** `lango economy escrow list` is run with escrow enabled
- **THEN** the command writes the escrow summary and detailed-show guidance to the Cobra command output stream

#### Scenario: Escrow list economy-disabled output writes to command output
- **WHEN** `lango economy escrow list` is run with economy disabled
- **THEN** the command writes the economy-disabled message to the Cobra command output stream

#### Scenario: Escrow list escrow-disabled output writes to command output
- **WHEN** `lango economy escrow list` is run with escrow disabled
- **THEN** the command writes the escrow-disabled message to the Cobra command output stream
