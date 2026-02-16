## Purpose

Define the configuration loading, saving, and migration system for encrypted SQLite profiles.
## Requirements
### Requirement: Configuration loading
The system SHALL load configuration through the bootstrap process from an encrypted SQLite database profile instead of directly from a plaintext JSON file. The `config.Load()` function SHALL be retained for migration purposes only.

#### Scenario: Normal startup
- **WHEN** the application starts via `lango serve`
- **THEN** configuration is loaded via `bootstrap.Run()` which reads the active encrypted profile

#### Scenario: Migration loading
- **WHEN** `config.Load()` is called during JSON import
- **THEN** the JSON file is read with environment variable substitution (existing behavior preserved)

### Requirement: Configuration save
The system SHALL save configuration through `configstore.Store.Save()` which encrypts and stores in the database. The legacy `config.Save()` function SHALL be removed.

#### Scenario: Save via configstore
- **WHEN** a config is saved through the configstore
- **THEN** it is JSON-serialized, AES-256-GCM encrypted, and stored in the database

#### Scenario: No legacy save function
- **WHEN** code attempts to call `config.Save()`
- **THEN** a compile error SHALL occur because the function no longer exists

### Requirement: Environment variable substitution
The system SHALL substitute environment variables in configuration values.

#### Scenario: Environment variable in value
- **WHEN** a config value contains ${VAR_NAME}
- **THEN** it SHALL be replaced with the environment variable value

#### Scenario: Missing environment variable
- **WHEN** a referenced environment variable is not set
- **THEN** an error SHALL be logged and default used if available

### Requirement: Configuration validation
The configuration system SHALL validate that at least one provider is configured with a non-empty `apiKey` or valid OAuth token. It SHALL validate that `agent.provider` references an existing key in the `providers` map. It SHALL NOT require `agent.apiKey` (this field no longer exists).

#### Scenario: Valid configuration
- **WHEN** config has `agent.provider: "google"` and `providers.google.type: "gemini"` with a valid `apiKey`
- **THEN** validation SHALL pass

#### Scenario: Invalid configuration
- **WHEN** config has `agent.provider: "google"` but no `google` key in `providers` map
- **THEN** validation SHALL fail with a clear error message

### Requirement: Default values
The configuration system SHALL apply sensible defaults for all non-credential fields. The minimum viable configuration SHALL require only: `agent.provider`, `providers.<name>.type`, `providers.<name>.apiKey`, and one channel's `enabled: true` + token. All other fields SHALL have defaults:
- `server.host`: `"localhost"`
- `server.port`: `18789`
- `server.httpEnabled`: `true`
- `server.wsEnabled`: `true`
- `session.databasePath`: `"~/.lango/data.db"`
- `session.maxHistoryTurns`: `100`
- `logging.level`: `"info"`
- `logging.format`: `"console"`
- `agent.maxTokens`: `4096`
- `agent.temperature`: `0.7`
- `tools.exec.defaultTimeout`: `30s`
- `tools.exec.allowBackground`: `true`
- `tools.filesystem.maxReadSize`: `1048576` (1MB)
- `tools.browser.headless`: `true`
- `tools.browser.sessionTimeout`: `5m`

#### Scenario: Missing optional field
- **WHEN** a configuration field is not specified
- **THEN** the system SHALL use the default value listed above
- **THEN** no error or warning SHALL be emitted for missing optional fields

#### Scenario: Minimal configuration startup
- **WHEN** config contains only `agent.provider`, one provider entry with `type` and `apiKey`, and one channel with `enabled: true` and token
- **THEN** the application SHALL start successfully with all defaults applied

### Requirement: Runtime configuration updates
The system SHALL support reloading configuration without full restart.

#### Scenario: Config file change
- **WHEN** the configuration file is modified
- **THEN** the system MAY reload affected components

#### Scenario: API config update
- **WHEN** configuration is updated via the Gateway API
- **THEN** the changes SHALL take effect for new operations

### Requirement: Providers Configuration Section
The system SHALL support a `providers` section in the configuration file to define multiple AI providers.

#### Scenario: Provider specific settings
- **WHEN** `providers` map is present in config
- **THEN** it SHALL map provider IDs (e.g., "openai", "anthropic") to their specific settings
- **AND** settings SHALL include `apiKey`, `baseUrl`, and provider-specific fields

#### Scenario: Fallback configuration
- **WHEN** `agent.fallbacks` list is present
- **THEN** it SHALL define an ordered list of fallback models
- **AND** each fallback SHALL specify `provider` and `model`

### Requirement: Provider Selection
The system SHALL allow selecting the active provider and model.

#### Scenario: Explicit provider selection
- **WHEN** `agent.provider` is set in config
- **THEN** the system SHALL use that provider for agent operations

#### Scenario: Default provider
- **WHEN** `agent.provider` is missing but `providers` has entries
- **THEN** the system SHALL adhere to a documented default behavior or return an error if ambiguous

### Requirement: Knowledge Configuration Section
The system SHALL support a `knowledge` section in the configuration for self-learning settings.

#### Scenario: Knowledge config fields
- **WHEN** `knowledge` section is present in configuration
- **THEN** it SHALL support the following fields:
  - `enabled` (bool): Enable the knowledge/learning system (default: false)
  - `maxLearnings` (int): Maximum learning entries per session (default: 10)
  - `maxKnowledge` (int): Maximum knowledge entries per session (default: 20)
  - `maxContextPerLayer` (int): Maximum context items per layer in retrieval (default: 5)
  - `autoApproveSkills` (bool): Auto-approve new skills without human review (default: false)
  - `maxSkillsPerDay` (int): Maximum new skills per day

#### Scenario: Knowledge disabled by default
- **WHEN** `knowledge` section is omitted from configuration
- **THEN** the system SHALL treat knowledge as disabled
- **AND** no knowledge-related initialization SHALL occur

#### Scenario: Knowledge config validation
- **WHEN** `knowledge.enabled` is true
- **THEN** the system SHALL apply default values for any omitted numeric fields
- **AND** `maxLearnings` SHALL default to 10 if not specified or <= 0
- **AND** `maxKnowledge` SHALL default to 20 if not specified or <= 0
- **AND** `maxContextPerLayer` SHALL default to 5 if not specified or <= 0

### Requirement: Graph config defaults
DefaultConfig SHALL include Graph defaults: Enabled=false, Backend="bolt", MaxTraversalDepth=2, MaxExpansionResults=10. Viper defaults SHALL be registered for these fields.

#### Scenario: New profile defaults
- **WHEN** a new profile is created via `lango config create`
- **THEN** graph config has Enabled=false, Backend="bolt", MaxTraversalDepth=2, MaxExpansionResults=10

### Requirement: A2A config defaults
DefaultConfig SHALL include A2A defaults: Enabled=false. Viper defaults SHALL be registered.

#### Scenario: New profile A2A defaults
- **WHEN** a new profile is created via `lango config create`
- **THEN** A2A config has Enabled=false

### Requirement: Graph config validation
Validate SHALL reject configurations where graph.enabled is true and graph.backend is not "bolt".

#### Scenario: Invalid graph backend
- **WHEN** config has graph.enabled=true and graph.backend="rocksdb"
- **THEN** Validate returns an error about unsupported backend

### Requirement: A2A config validation
Validate SHALL reject configurations where a2a.enabled is true but a2a.baseUrl or a2a.agentName is empty.

#### Scenario: A2A missing base URL
- **WHEN** config has a2a.enabled=true and a2a.baseUrl is empty
- **THEN** Validate returns an error about required baseUrl

#### Scenario: A2A missing agent name
- **WHEN** config has a2a.enabled=true and a2a.agentName is empty
- **THEN** Validate returns an error about required agentName

