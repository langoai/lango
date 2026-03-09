## ADDED Requirements

### Requirement: On-chain escrow documentation in economy.md
The system SHALL include documentation for on-chain escrow (Hub/Vault dual-mode) in `docs/features/economy.md`, covering deal lifecycle, contract architecture, and configuration.

#### Scenario: Hub vs Vault mode documentation
- **WHEN** a user reads the on-chain escrow section in economy.md
- **THEN** they find descriptions of Hub mode (single contract, multiple deals) and Vault mode (per-deal EIP-1167 proxy)

#### Scenario: On-chain config keys in configuration.md
- **WHEN** a user reads `docs/configuration.md`
- **THEN** all 10 on-chain escrow config keys (`economy.escrow.onChain.*`, `economy.escrow.settlement.*`) are documented with types and defaults

### Requirement: On-chain escrow CLI documentation
The system SHALL document `escrow list`, `escrow show`, and `escrow sentinel status` CLI commands in `docs/cli/economy.md`.

#### Scenario: CLI command reference
- **WHEN** a user reads `docs/cli/economy.md`
- **THEN** they find usage, flags, and output examples for `lango economy escrow list`, `lango economy escrow show`, and `lango economy escrow sentinel status`

### Requirement: Escrow tools in system prompts
The system SHALL list all 10 `escrow_*` tools with correct names and workflow guidance in `prompts/TOOL_USAGE.md`.

#### Scenario: Tool names match code
- **WHEN** the agent reads TOOL_USAGE.md
- **THEN** tool names match those registered in `internal/app/tools_escrow.go`: `escrow_create`, `escrow_fund`, `escrow_activate`, `escrow_submit_work`, `escrow_release`, `escrow_refund`, `escrow_dispute`, `escrow_resolve`, `escrow_status`, `escrow_list`

### Requirement: Contracts documentation
The system SHALL document Foundry-based escrow contracts (LangoEscrowHub, LangoVault, LangoVaultFactory) in `docs/features/contracts.md`.

#### Scenario: Contract architecture documented
- **WHEN** a user reads `docs/features/contracts.md`
- **THEN** they find contract descriptions, deal states, events, and Foundry build/test commands

### Requirement: On-chain escrow events in economy.md
The system SHALL document the 6 new on-chain events in the Events Summary table of `docs/features/economy.md`.

#### Scenario: Events table updated
- **WHEN** a user reads the Events Summary in economy.md
- **THEN** events for DealCreated, DealDeposited, WorkSubmitted, DealReleased, DealRefunded, DealDisputed are listed

### Requirement: README reflects on-chain escrow
The system SHALL mention on-chain Hub/Vault escrow, Foundry contracts, and escrow CLI commands in `README.md`.

#### Scenario: Feature bullets updated
- **WHEN** a user reads README.md features section
- **THEN** on-chain escrow and Foundry contracts are mentioned
