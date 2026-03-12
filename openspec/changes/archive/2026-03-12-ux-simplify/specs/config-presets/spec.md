## ADDED Requirements

### Requirement: Preset profile creation
The system SHALL provide `PresetConfig(name)` that returns a Config with feature flags set according to the named preset. Unknown names SHALL return DefaultConfig().

#### Scenario: Minimal preset
- **WHEN** PresetConfig("minimal") is called
- **THEN** returned config matches DefaultConfig()

#### Scenario: Researcher preset
- **WHEN** PresetConfig("researcher") is called
- **THEN** Knowledge, ObservationalMemory, Graph, Librarian are enabled, Embedding provider is "openai" with model "text-embedding-3-small"

#### Scenario: Collaborator preset
- **WHEN** PresetConfig("collaborator") is called
- **THEN** P2P, Payment, Economy are enabled; Knowledge features remain disabled

#### Scenario: Full preset
- **WHEN** PresetConfig("full") is called
- **THEN** Knowledge, ObservationalMemory, Graph, Librarian, Cron, Background, Workflow, MCP, AgentMemory, MultiAgent are all enabled

### Requirement: Preset validation
The system SHALL provide `IsValidPreset(name)` returning true for "minimal", "researcher", "collaborator", "full" and false otherwise.

#### Scenario: Valid preset name
- **WHEN** IsValidPreset("researcher") is called
- **THEN** returns true

#### Scenario: Invalid preset name
- **WHEN** IsValidPreset("unknown") is called
- **THEN** returns false

### Requirement: CLI preset flag
The `lango config create` command SHALL accept a `--preset` flag that uses PresetConfig() instead of DefaultConfig() when creating a profile.

#### Scenario: Create with preset
- **WHEN** user runs `lango config create my-bot --preset researcher`
- **THEN** profile "my-bot" is created with researcher preset configuration

#### Scenario: Invalid preset error
- **WHEN** user runs `lango config create my-bot --preset invalid`
- **THEN** command returns an error listing valid preset names
