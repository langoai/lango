## ADDED Requirements

### Requirement: Escrow CLI show output routing

`lango economy escrow show` SHALL write all non-error output through the Cobra command output stream so wrappers and test harnesses can capture detailed-config, disabled-state, and ID-guidance output without intercepting process-global stdout.

#### Scenario: Escrow show detailed config writes to command output
- **WHEN** `lango economy escrow show` is run without `--id`
- **THEN** the command writes the detailed on-chain escrow configuration and settlement section to the Cobra command output stream

#### Scenario: Escrow show disabled output writes to command output
- **WHEN** `lango economy escrow show` is run with escrow disabled
- **THEN** the command writes the disabled-state message to the Cobra command output stream

#### Scenario: Escrow show ID guidance writes to command output
- **WHEN** `lango economy escrow show --id escrow-123` is run
- **THEN** the command writes the live-status runtime guidance to the Cobra command output stream
