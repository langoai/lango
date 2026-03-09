## ADDED Requirements

### Requirement: Settlement documentation
The system SHALL document P2P settlement workflow in `docs/features/economy.md`, covering settlement config keys and receipt confirmation flow.

#### Scenario: Settlement config documented
- **WHEN** a user reads the on-chain escrow section in economy.md
- **THEN** they find `economy.escrow.settlement.receiptTimeout` and `economy.escrow.settlement.maxRetries` documented

#### Scenario: Settlement in configuration.md
- **WHEN** a user reads `docs/configuration.md`
- **THEN** settlement config keys are listed in the escrow configuration table
