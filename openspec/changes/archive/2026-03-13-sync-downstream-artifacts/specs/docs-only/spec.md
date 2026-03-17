## MODIFIED Requirements

### Requirement: Quickstart references config presets
The getting started quickstart documentation SHALL reference the `--preset` flag and link to the config presets documentation.

#### Scenario: Preset flag in quickstart
- **WHEN** a user reads `docs/getting-started/quickstart.md`
- **THEN** the `--preset` flag SHALL be mentioned with a brief preset table and link to `config-presets.md`

### Requirement: CLI index includes status command
The CLI index quick reference table SHALL include the `lango status` command.

#### Scenario: Status in CLI index
- **WHEN** a user reads `docs/cli/index.md`
- **THEN** `lango status` SHALL appear in the Quick Reference table under Getting Started
