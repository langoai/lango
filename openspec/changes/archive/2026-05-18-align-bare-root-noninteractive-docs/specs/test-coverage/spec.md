## ADDED Requirements

### Requirement: Docs guard covers bare root non-interactive contract
Executable docs quality coverage SHALL fail when public CLI docs omit the bare-root non-interactive help fallback contract.

#### Scenario: Public docs guard checks bare root fallback
- **WHEN** docs quality tests run
- **THEN** README, `docs/cli/index.md`, and `docs/cli/core.md` SHALL be checked for the interactive bare-root launch contract
- **AND** they SHALL be checked for the non-interactive help fallback contract
- **AND** they SHALL be checked for the distinction from `lango cockpit` and `lango chat` non-interactive behavior
