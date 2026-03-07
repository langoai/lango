## ADDED Requirements

### Requirement: Contract feature documentation page
The documentation site SHALL include a `docs/features/contracts.md` page documenting smart contract interaction capabilities including ABI cache, read (view/pure), and write (state-changing) operations, with experimental warning, architecture overview, agent tools listing, and configuration reference.

#### Scenario: Contract feature docs page exists
- **WHEN** the documentation site is built
- **THEN** `docs/features/contracts.md` SHALL exist with sections for ABI cache, read operations, write operations, agent tools, and configuration

### Requirement: Contract CLI documentation page
The documentation site SHALL include a `docs/cli/contract.md` page documenting `lango contract read`, `lango contract call`, and `lango contract abi load` commands with flags tables and example output following the `docs/cli/payment.md` pattern.

#### Scenario: Contract CLI docs page exists
- **WHEN** the documentation site is built
- **THEN** `docs/cli/contract.md` SHALL exist with sections for read, call, and abi load subcommands

#### Scenario: Each subcommand documented with flags
- **WHEN** a user reads the contract CLI reference
- **THEN** each subcommand SHALL include a flags table with `--address`, `--abi`, `--method`, `--args`, `--chain-id`, and `--output` flags documented
