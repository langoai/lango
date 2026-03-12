## MODIFIED Requirements

### Requirement: Onboard preset support
The onboard command SHALL accept a `--preset` flag (minimal, researcher, collaborator, full) that initializes the wizard config from the named preset instead of DefaultConfig().

#### Scenario: Onboard with preset
- **WHEN** user runs `lango onboard --preset researcher`
- **THEN** wizard starts with researcher preset config (Knowledge, Graph, etc. pre-enabled)

#### Scenario: Invalid preset
- **WHEN** user runs `lango onboard --preset invalid`
- **THEN** command returns error listing valid presets

### Requirement: Config-aware next steps
After onboard completion, the system SHALL display recommended features that are currently disabled, with the settings category name for each.

#### Scenario: Default config recommendations
- **WHEN** onboard completes with default config (Knowledge, ObsMemory, Cron, MCP all disabled)
- **THEN** next steps shows all four as recommendations with their settings category names

#### Scenario: Researcher preset recommendations
- **WHEN** onboard completes with researcher preset (Knowledge enabled)
- **THEN** Knowledge is NOT listed in recommendations (already enabled); Cron and MCP are listed

### Requirement: Preset hints in next steps
The next steps output SHALL include quick preset commands for creating additional profiles.

#### Scenario: Preset commands shown
- **WHEN** onboard completes
- **THEN** output includes example commands: `lango config create <name> --preset researcher/collaborator/full`
