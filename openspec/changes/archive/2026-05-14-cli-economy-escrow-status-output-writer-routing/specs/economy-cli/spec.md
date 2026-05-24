## ADDED Requirements

### Requirement: Escrow CLI status output routing

`lango economy escrow status` SHALL write all non-error output through the Cobra command output stream so wrappers and test harnesses can capture enabled and disabled output without intercepting process-global stdout.

#### Scenario: Escrow status enabled output writes to command output
- **WHEN** `lango economy escrow status` is run with escrow enabled
- **THEN** the command writes the escrow configuration to the Cobra command output stream

#### Scenario: Escrow status disabled output writes to command output
- **WHEN** `lango economy escrow status` is run with escrow disabled
- **THEN** the command writes the disabled-state message to the Cobra command output stream
