## ADDED Requirements

### Requirement: Economy CLI documentation page
The documentation site SHALL include a `docs/cli/economy.md` page documenting all economy CLI commands with subcommand sections, flags tables, and example output following the `docs/cli/payment.md` pattern.

#### Scenario: Economy CLI docs page exists
- **WHEN** the documentation site is built
- **THEN** `docs/cli/economy.md` SHALL exist with sections for budget, risk, pricing, negotiate, and escrow subcommands

#### Scenario: Each subcommand documented with flags and output
- **WHEN** a user reads the economy CLI reference
- **THEN** each subcommand section SHALL include a flags table (if applicable) and example terminal output
