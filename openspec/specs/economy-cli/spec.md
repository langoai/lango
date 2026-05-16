## Purpose

CLI commands for managing and inspecting the P2P economy layer subsystems (budget, risk, pricing, negotiation, escrow).

## Requirements

### Requirement: Economy CLI command group
The system SHALL provide a `lango economy` CLI command group with subcommands for budget, risk, pricing, negotiate, and escrow. The command group SHALL be registered under GroupID "infra".

#### Scenario: Economy help
- **WHEN** `lango economy --help` is run
- **THEN** all 5 subcommands are listed with descriptions

#### Scenario: Economy help examples use current commands
- **WHEN** `lango economy --help` is run
- **THEN** its examples SHALL use the current status/show surface
- **AND** it SHALL NOT mention nonexistent examples such as `risk assess`, `pricing quote`, `negotiate list`, or `escrow status --escrow-id`

### Requirement: Budget CLI
The system SHALL provide `lango economy budget` that displays budget subsystem status including enabled state and configuration.

#### Scenario: Budget status
- **WHEN** `lango economy budget` is run
- **THEN** budget configuration (defaultMax, hardLimit, alertThresholds) is displayed

### Requirement: Budget CLI output routing
`lango economy budget` SHALL write all non-error output through the Cobra command output stream so wrappers and test harnesses can capture enabled, disabled, and task-guidance output without intercepting process-global stdout.

#### Scenario: Budget enabled output writes to command output
- **WHEN** `lango economy budget` is run with economy enabled
- **THEN** the command writes the budget configuration to the Cobra command output stream

#### Scenario: Budget task guidance writes to command output
- **WHEN** `lango economy budget --task-id=task-1` is run
- **THEN** the command writes the task guidance to the Cobra command output stream

#### Scenario: Budget disabled output writes to command output
- **WHEN** `lango economy budget` is run with economy disabled
- **THEN** the command writes the disabled-state message to the Cobra command output stream

### Requirement: Risk CLI
The system SHALL provide `lango economy risk` that displays risk assessment subsystem status including configuration and strategy matrix.

#### Scenario: Risk status
- **WHEN** `lango economy risk` is run
- **THEN** risk configuration (escrowThreshold, factor weights) is displayed

### Requirement: Risk CLI output routing
`lango economy risk` SHALL write all non-error output through the Cobra command output stream so wrappers and test harnesses can capture enabled and disabled output without intercepting process-global stdout.

#### Scenario: Risk enabled output writes to command output
- **WHEN** `lango economy risk` is run with economy enabled
- **THEN** the command writes the risk configuration to the Cobra command output stream

#### Scenario: Risk disabled output writes to command output
- **WHEN** `lango economy risk` is run with economy disabled
- **THEN** the command writes the disabled-state message to the Cobra command output stream

### Requirement: Pricing CLI
The system SHALL provide `lango economy pricing` that displays dynamic pricing subsystem status including base prices and discount configuration.

#### Scenario: Pricing status
- **WHEN** `lango economy pricing` is run
- **THEN** pricing configuration (basePrices, trustDiscount, volumeDiscount) is displayed

### Requirement: Pricing CLI output routing
`lango economy pricing` SHALL write all non-error output through the Cobra command output stream so wrappers and test harnesses can capture enabled and disabled output without intercepting process-global stdout.

#### Scenario: Pricing enabled output writes to command output
- **WHEN** `lango economy pricing` is run with pricing enabled
- **THEN** the command writes the pricing configuration to the Cobra command output stream

#### Scenario: Pricing disabled output writes to command output
- **WHEN** `lango economy pricing` is run with pricing disabled
- **THEN** the command writes the disabled-state message to the Cobra command output stream

### Requirement: Negotiate CLI
The system SHALL provide `lango economy negotiate` that displays negotiation subsystem status including session timeout and max rounds.

#### Scenario: Negotiate status
- **WHEN** `lango economy negotiate` is run
- **THEN** negotiation configuration (maxRounds, sessionTimeout) is displayed

### Requirement: Negotiate CLI output routing
`lango economy negotiate` SHALL write all non-error output through the Cobra command output stream so wrappers and test harnesses can capture enabled and disabled output without intercepting process-global stdout.

#### Scenario: Negotiate enabled output writes to command output
- **WHEN** `lango economy negotiate` is run with negotiation enabled
- **THEN** the command writes the negotiation configuration to the Cobra command output stream

#### Scenario: Negotiate disabled output writes to command output
- **WHEN** `lango economy negotiate` is run with negotiation disabled
- **THEN** the command writes the disabled-state message to the Cobra command output stream

### Requirement: Escrow CLI
The system SHALL provide `lango economy escrow` that displays escrow subsystem status including timeout and settlement configuration.

#### Scenario: Escrow status
- **WHEN** `lango economy escrow` is run
- **THEN** escrow configuration (timeout, maxMilestones) is displayed

### Requirement: Escrow CLI status output routing
`lango economy escrow status` SHALL write all non-error output through the Cobra command output stream so wrappers and test harnesses can capture enabled and disabled output without intercepting process-global stdout.

#### Scenario: Escrow status enabled output writes to command output
- **WHEN** `lango economy escrow status` is run with escrow enabled
- **THEN** the command writes the escrow configuration to the Cobra command output stream

#### Scenario: Escrow status disabled output writes to command output
- **WHEN** `lango economy escrow status` is run with escrow disabled
- **THEN** the command writes the disabled-state message to the Cobra command output stream

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

### Requirement: Economy CLI documentation page
The documentation site SHALL include a `docs/cli/economy.md` page documenting all economy CLI commands with subcommand sections, flags tables, and example output following the `docs/cli/payment.md` pattern.

#### Scenario: Economy CLI docs page exists
- **WHEN** the documentation site is built
- **THEN** `docs/cli/economy.md` SHALL exist with sections for budget, risk, pricing, negotiate, and escrow subcommands

#### Scenario: Each subcommand documented with flags and output
- **WHEN** a user reads the economy CLI reference
- **THEN** each subcommand section SHALL include a flags table (if applicable) and example terminal output
