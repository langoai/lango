## ADDED Requirements

### Requirement: Economy CLI command group
The system SHALL provide a `lango economy` CLI command group with subcommands for budget, risk, pricing, negotiate, and escrow. The command group SHALL be registered under GroupID "infra".

#### Scenario: Economy help
- **WHEN** `lango economy --help` is run
- **THEN** all 5 subcommands are listed with descriptions

### Requirement: Budget CLI
The system SHALL provide `lango economy budget` that displays budget subsystem status including enabled state and configuration.

#### Scenario: Budget status
- **WHEN** `lango economy budget` is run
- **THEN** budget configuration (defaultMax, hardLimit, alertThresholds) is displayed

### Requirement: Risk CLI
The system SHALL provide `lango economy risk` that displays risk assessment subsystem status including configuration and strategy matrix.

#### Scenario: Risk status
- **WHEN** `lango economy risk` is run
- **THEN** risk configuration (escrowThreshold, factor weights) is displayed

### Requirement: Pricing CLI
The system SHALL provide `lango economy pricing` that displays dynamic pricing subsystem status including base prices and discount configuration.

#### Scenario: Pricing status
- **WHEN** `lango economy pricing` is run
- **THEN** pricing configuration (basePrices, trustDiscount, volumeDiscount) is displayed

### Requirement: Negotiate CLI
The system SHALL provide `lango economy negotiate` that displays negotiation subsystem status including session timeout and max rounds.

#### Scenario: Negotiate status
- **WHEN** `lango economy negotiate` is run
- **THEN** negotiation configuration (maxRounds, sessionTimeout) is displayed

### Requirement: Escrow CLI
The system SHALL provide `lango economy escrow` that displays escrow subsystem status including timeout and settlement configuration.

#### Scenario: Escrow status
- **WHEN** `lango economy escrow` is run
- **THEN** escrow configuration (timeout, maxMilestones) is displayed
