## Purpose

Define the configuration loading, saving, and migration system for encrypted SQLite profiles.
## Requirements
### Requirement: PostLoad one-stop normalization and validation
The `config` package SHALL export a `PostLoad(*Config) error` function that applies all post-load processing in order: legacy migration, environment variable substitution, path normalization, path validation, and full config validation. All operations MUST be idempotent — calling PostLoad multiple times on the same config SHALL produce the same result.

#### Scenario: PostLoad applies full processing chain
- **WHEN** `PostLoad(cfg)` is called on a freshly deserialized config
- **THEN** the config has legacy fields migrated, env vars expanded, paths normalized to absolute, data paths validated under DataRoot, and full config validation applied

#### Scenario: PostLoad is idempotent
- **WHEN** `PostLoad(cfg)` is called twice on the same config
- **THEN** the second call produces no additional changes and returns the same result

### Requirement: Configuration loading
The system SHALL load configuration through the bootstrap process from an encrypted SQLite database profile instead of directly from a plaintext JSON file. The `config.Load()` function SHALL be retained for migration purposes only. `Load()` SHALL return `(*LoadResult, error)` containing both the `*Config` and `ExplicitKeys map[string]bool`. `Load()` SHALL delegate all post-load processing to `PostLoad()` instead of calling individual steps separately. All existing callers of `config.Load()` MUST use `result.Config` for the config object.

#### Scenario: Normal startup
- **WHEN** the application starts via `lango serve`
- **THEN** configuration is loaded via `bootstrap.Run()` which reads the active encrypted profile

#### Scenario: Migration loading
- **WHEN** `config.Load()` is called during JSON import
- **THEN** the JSON file is read with environment variable substitution (existing behavior preserved)

#### Scenario: Load delegates to PostLoad
- **WHEN** `config.Load(path)` is called
- **THEN** after unmarshalling, it calls `ApplyContextProfile()` then `PostLoad(cfg)` once and returns the LoadResult

#### Scenario: Existing callers compile after signature change
- **WHEN** `config.Load()` is called from `internal/configstore/migrate.go` or other callers
- **THEN** each caller accesses `result.Config` and the project builds without errors

#### Scenario: ExplicitKeys collected from raw viper
- **WHEN** config file sets `knowledge.enabled: true` and `librarian.enabled: false`
- **THEN** `LoadResult.ExplicitKeys` contains both keys, and does NOT contain keys only present via `SetDefault()`

### Requirement: ContextProfile field in Config
`Config` SHALL have a `ContextProfile ContextProfileName` field with mapstructure tag `contextProfile`. The field SHALL accept values `off`, `lite`, `balanced`, `full`, or empty string (no profile).

#### Scenario: ContextProfile unmarshaled from JSON config
- **WHEN** config file contains `"contextProfile": "balanced"`
- **THEN** `cfg.ContextProfile` equals `ContextProfileBalanced`

#### Scenario: Empty profile means no profile applied
- **WHEN** config file does not set `contextProfile`
- **THEN** `cfg.ContextProfile` is empty string and `ApplyContextProfile` is a no-op

### Requirement: ApplyContextProfile in load pipeline
`ApplyContextProfile(cfg, explicitKeys)` SHALL be called inside `Load()` after `Unmarshal` and before `PostLoad`. `PostLoad`'s signature SHALL NOT change.

#### Scenario: Profile applied before validation
- **WHEN** `contextProfile: balanced` is set
- **THEN** `Knowledge.Enabled` is `true` before `Validate()` runs, so downstream validation sees the profile-applied state

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

### Requirement: Fallback provider existence validation
`config.Validate()` SHALL verify that `agent.fallbackProvider` (when set) references an existing key in the `providers` map.

#### Scenario: Fallback provider not in providers map
- **WHEN** `agent.fallbackProvider` is set to a value not present in the `providers` map
- **THEN** validation SHALL fail with an error identifying the missing provider

#### Scenario: Fallback provider exists
- **WHEN** `agent.fallbackProvider` references a valid key in the `providers` map
- **THEN** validation SHALL pass (no error for this check)

### Requirement: Provider-model compatibility validation at startup
`config.Validate()` SHALL check both primary (`agent.provider`/`agent.model`) and fallback (`agent.fallbackProvider`/`agent.fallbackModel`) pairs for model-provider compatibility using `ValidateModelProvider`.

#### Scenario: Primary model incompatible with provider type
- **WHEN** `agent.model` is `gpt-5.3-codex` and `agent.provider` references a gemini-type provider
- **THEN** validation SHALL fail with an error describing the mismatch

#### Scenario: Fallback model incompatible with fallback provider type
- **WHEN** `agent.fallbackModel` is `claude-sonnet-4-5-20250514` and `agent.fallbackProvider` references an openai-type provider
- **THEN** validation SHALL fail with an error describing the mismatch

### Requirement: Configuration validation
The configuration system SHALL validate that at least one provider is configured with a non-empty `apiKey` or valid OAuth token. It SHALL validate that `agent.provider` references an existing key in the `providers` map. It SHALL NOT require `agent.apiKey` (this field no longer exists). The `Validate()` function SHALL reference exported package-level validation maps (`ValidLogLevels`, `ValidLogFormats`, `ValidSignerProviders`, `ValidWalletProviders`, `ValidZKPSchemes`, `ValidContainerRuntimes`, `ValidMCPTransports`) from `config/constants.go` instead of inline map literals.

#### Scenario: Valid configuration
- **WHEN** config has `agent.provider: "google"` and `providers.google.type: "gemini"` with a valid `apiKey`
- **THEN** validation SHALL pass

#### Scenario: Invalid configuration
- **WHEN** config has `agent.provider: "google"` but no `google` key in `providers` map
- **THEN** validation SHALL fail with a clear error message

#### Scenario: Validation map reuse
- **WHEN** `config.Validate()` checks the log level value
- **THEN** it uses `config.ValidLogLevels` map defined in `constants.go`

#### Scenario: External access to valid values
- **WHEN** another package needs to validate a config value (e.g., CLI flag validation)
- **THEN** it can import and use `config.ValidLogLevels` directly

### Requirement: Default values
DefaultConfig() SHALL be the single source of truth for all config default values. Load() SHALL derive all viper defaults by walking the DefaultConfig() struct recursively using mapstructure tags and calling v.SetDefault() for each non-zero leaf field. Manual v.SetDefault() calls for individual config keys SHALL NOT exist in the loading path.

#### Scenario: Load uses struct walker for defaults
- **WHEN** `config.Load(path)` is called
- **THEN** it SHALL walk `DefaultConfig()` struct via `setDefaultsFromStruct()` to populate all viper defaults

#### Scenario: New config fields are automatically defaulted
- **WHEN** a developer adds a new field with a mapstructure tag and non-zero default in DefaultConfig()
- **THEN** Load() SHALL apply that default automatically without manual SetDefault calls

#### Scenario: No manual SetDefault in load path
- **WHEN** the Load() function is inspected
- **THEN** there SHALL be zero manual v.SetDefault() calls outside the walker

#### Scenario: Parity between DefaultConfig and viper unmarshal
- **WHEN** DefaultConfig() is compared with a Config produced by viper unmarshal using only walker-derived defaults
- **THEN** all non-zero fields SHALL match

### Requirement: DataRoot enforces data path boundaries
The Config SHALL include a `DataRoot` field (default: `~/.lango/`) that defines the root directory for all lango data files. All configurable data paths (session.databasePath, graph.databasePath, skill.skillsDir, workflow.stateDir, p2p.keyDir, p2p.zkp.proofCacheDir, p2p.workspace.dataDir) MUST reside under DataRoot. The `NormalizePaths()` function SHALL expand tildes and resolve relative paths under DataRoot. The `ValidateDataPaths()` function SHALL reject any path outside DataRoot with a clear error message.

#### Scenario: Default paths pass validation
- **WHEN** config uses default paths (all under ~/.lango/)
- **THEN** NormalizePaths and ValidateDataPaths succeed

#### Scenario: External path rejected
- **WHEN** graph.databasePath is set to "/tmp/graph.db"
- **THEN** ValidateDataPaths returns error "graph.databasePath must be under data root"

#### Scenario: Relative path resolved under DataRoot
- **WHEN** graph.databasePath is set to "graph.db" (relative)
- **THEN** NormalizePaths resolves it to `<DataRoot>/graph.db`

#### Scenario: Custom DataRoot accepted
- **WHEN** DataRoot is set to "/data/lango" and all paths are under it
- **THEN** validation passes

### Requirement: ExecToolConfig supports additional protected paths
The ExecToolConfig SHALL include an `AdditionalProtectedPaths` field that specifies extra paths for CommandGuard to protect, in addition to DataRoot.

#### Scenario: Additional path protected
- **WHEN** additionalProtectedPaths includes "/var/secrets"
- **THEN** exec commands accessing /var/secrets are blocked

### Requirement: ExpandEnvVars is exported
The `config` package SHALL export `ExpandEnvVars(s string) string` as a public function that replaces `${VAR}` patterns with environment variable values. Variables not set in the environment SHALL be left as-is.

#### Scenario: Env var expansion from external package
- **WHEN** `config.ExpandEnvVars("${OPENAI_API_KEY}")` is called and `OPENAI_API_KEY` is set
- **THEN** the function SHALL return the environment variable value

#### Scenario: Unset env var preserved
- **WHEN** `config.ExpandEnvVars("${UNSET_VAR}")` is called and `UNSET_VAR` is not set
- **THEN** the function SHALL return `"${UNSET_VAR}"` unchanged

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

### Requirement: Configuration structure — Payment
The Config struct SHALL include a `Payment PaymentConfig` field after the A2A field. PaymentConfig SHALL contain: Enabled (bool), WalletProvider (string: local/rpc/composite), Network (PaymentNetworkConfig), Limits (SpendingLimitsConfig), X402 (X402Config).

PaymentNetworkConfig SHALL contain: ChainID (int64, default 84532), RPCURL (string), USDCContract (string).
SpendingLimitsConfig SHALL contain: MaxPerTx (string, default "1.00"), MaxDaily (string, default "10.00"), AutoApproveBelow (string).
X402Config SHALL contain: AutoIntercept (bool), MaxAutoPayAmount (string).

#### Scenario: Default payment config
- **WHEN** no payment config is specified
- **THEN** payment is disabled with Base Sepolia defaults (chainId 84532, maxPerTx "1.00", maxDaily "10.00")

### Requirement: Payment config validation
The Validate function SHALL check: when payment.enabled is true, payment.network.rpcUrl MUST be non-empty. payment.walletProvider MUST be one of "local", "rpc", or "composite".

#### Scenario: Payment enabled without RPC URL
- **WHEN** payment.enabled is true and rpcUrl is empty
- **THEN** validation fails with an error message

### Requirement: Payment environment variable substitution
The substituteEnvVars function SHALL expand `${VAR}` patterns in `payment.network.rpcUrl`.

#### Scenario: RPC URL from environment
- **WHEN** rpcUrl is set to `${BASE_RPC_URL}`
- **THEN** the environment variable value is substituted

### Requirement: Cron configuration
The config system SHALL support a `cron` section with fields: enabled (bool), timezone (string), maxConcurrentJobs (int), defaultSessionMode (string), historyRetention (duration string), defaultDeliverTo ([]string).

#### Scenario: Default cron config
- **WHEN** no cron config is specified
- **THEN** defaults SHALL be: enabled=false, timezone="UTC", maxConcurrentJobs=5, defaultSessionMode="isolated", historyRetention="720h", defaultDeliverTo=nil

### Requirement: Background configuration
The config system SHALL support a `background` section with fields: enabled (bool), yieldMs (int), maxConcurrentTasks (int), defaultDeliverTo ([]string).

#### Scenario: Default background config
- **WHEN** no background config is specified
- **THEN** defaults SHALL be: enabled=false, yieldMs=30000, maxConcurrentTasks=3, defaultDeliverTo=nil

### Requirement: Workflow configuration
The config system SHALL support a `workflow` section with fields: enabled (bool), maxConcurrentSteps (int), defaultTimeout (duration string), stateDir (string), defaultDeliverTo ([]string).

#### Scenario: Default workflow config
- **WHEN** no workflow config is specified
- **THEN** defaults SHALL be: enabled=false, maxConcurrentSteps=4, defaultTimeout="10m", stateDir="~/.lango/workflows/", defaultDeliverTo=nil

### Requirement: Automation config DefaultDeliverTo fields
CronConfig, BackgroundConfig, and WorkflowConfig SHALL each include a `DefaultDeliverTo []string` field with mapstructure tag "defaultDeliverTo". The config loader SHALL register viper defaults for all three fields.

#### Scenario: Default config values
- **WHEN** the application starts with no explicit defaultDeliverTo configuration
- **THEN** the DefaultDeliverTo fields SHALL default to nil (empty slice)

#### Scenario: Config file specifies defaults
- **WHEN** the config file sets cron.defaultDeliverTo to ["telegram"]
- **THEN** the loaded CronConfig.DefaultDeliverTo SHALL contain ["telegram"]

### Requirement: Example config includes PII and Presidio fields
The example `config.json` SHALL include `piiDisabledPatterns` (empty array), `piiCustomPatterns` (empty object), and a `presidio` block with `enabled`, `url`, `scoreThreshold`, and `language` fields within the `security.interceptor` section.

#### Scenario: Docker headless user imports example config
- **WHEN** a user copies config.json for Docker headless deployment
- **THEN** the interceptor block contains `piiDisabledPatterns`, `piiCustomPatterns`, and `presidio` fields with sensible defaults
- **THEN** `presidio.enabled` defaults to `false` and `presidio.url` defaults to `http://localhost:5002`

### Requirement: InterceptorConfig PII pattern fields
InterceptorConfig SHALL include PIIDisabledPatterns ([]string), PIICustomPatterns (map[string]string), and Presidio (PresidioConfig) fields with appropriate mapstructure and json tags.

#### Scenario: Disabled patterns config
- **WHEN** config JSON contains "piiDisabledPatterns": ["passport", "ipv4"]
- **THEN** InterceptorConfig.PIIDisabledPatterns SHALL be ["passport", "ipv4"]

#### Scenario: Custom patterns config
- **WHEN** config JSON contains "piiCustomPatterns": {"my_id": "\\bID-\\d+\\b"}
- **THEN** InterceptorConfig.PIICustomPatterns SHALL contain the mapping

### Requirement: PresidioConfig type
A new PresidioConfig struct SHALL define Enabled (bool), URL (string, default "http://localhost:5002"), ScoreThreshold (float64, default 0.7), and Language (string, default "en").

#### Scenario: Presidio config loading
- **WHEN** config JSON contains presidio block with enabled=true, url, scoreThreshold, language
- **THEN** InterceptorConfig.Presidio SHALL be populated

#### Scenario: Default values
- **WHEN** no Presidio config is specified
- **THEN** URL SHALL default to "http://localhost:5002", ScoreThreshold to 0.7, Language to "en"

### Requirement: Provenance Configuration Section
The Config struct SHALL include a `Provenance ProvenanceConfig` field with sub-struct `CheckpointConfig`. DefaultConfig SHALL set provenance defaults: enabled=false, autoOnStepComplete=true, autoOnPolicy=true, maxPerSession=100, retentionDays=30.

#### Scenario: Default config includes provenance
- **WHEN** DefaultConfig() is called
- **THEN** the Provenance field is populated with default values

