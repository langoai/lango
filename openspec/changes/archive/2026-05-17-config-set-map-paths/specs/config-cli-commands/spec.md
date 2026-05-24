## MODIFIED Requirements

### Requirement: Config set uses single bootstrap with cleanup
The `config set` command SHALL bootstrap exactly once. The cfgLoader function SHALL return a cleanup function that closes the DB client. The cleanup function MUST be called via `defer` in `RunE` to ensure resources are released on all code paths (success, setConfigPath error, save error).

`config set` SHALL preserve the loaded profile's `ExplicitKeys` metadata when saving. When the set path is one of `config.ContextRelatedKeys()`, the saved explicit-key map SHALL include that path so later context auto-enable resolution treats the value as user-explicit.

Invalid dot-path errors from `config set` SHALL include actionable key discovery help. When nearby valid keys exist, the error SHALL include up to three deterministic suggestions. The command SHALL NOT save after an invalid path.

`config set` SHALL support map-backed dot paths whose map keys are dynamic user names. When a map entry is missing and the remaining path is valid for the map value type, the command SHALL create the map and entry before setting the requested value. When the final map value is a `string`, the final path segment SHALL be treated as the map key and the provided value SHALL be stored as the map value.

#### Scenario: Successful set closes DB client
- **WHEN** `config set agent.provider openai` succeeds
- **THEN** the DB client is closed after the command completes

#### Scenario: setConfigPath error closes DB client
- **WHEN** `config set invalid.key value` fails at setConfigPath
- **THEN** the cleanup function is still called via defer, closing the DB client

#### Scenario: Save error closes DB client
- **WHEN** save fails (e.g., validation error from PostLoad in Save)
- **THEN** the cleanup function is still called via defer, closing the DB client

#### Scenario: Loader failure does not leak resources
- **WHEN** the cfgLoader fails (bootstrap error)
- **THEN** cleanup is nil, defer is a no-op, no DB client exists to leak

#### Scenario: Config get/set/keys output uses the command writer
- **WHEN** `lango config get`, `lango config set`, or `lango config keys` renders output
- **THEN** it SHALL write the full output through the Cobra command output writer
- **AND** wrappers or tests that replace `cmd.OutOrStdout()` SHALL capture the command output

#### Scenario: Existing explicit keys preserved
- **WHEN** `lango config set agent.provider openai` saves a profile whose `ExplicitKeys` includes `knowledge.enabled`
- **THEN** the saved profile SHALL still include `knowledge.enabled` in `ExplicitKeys`

#### Scenario: Context-related set marks explicit
- **WHEN** `lango config set knowledge.enabled false` saves a profile
- **THEN** the saved profile SHALL include `knowledge.enabled` in `ExplicitKeys`

#### Scenario: Invalid set path does not mutate explicit keys
- **WHEN** `lango config set invalid.key value` fails before save
- **THEN** the loaded explicit-key map SHALL NOT be mutated

#### Scenario: Config set suggests nearby key on invalid path
- **WHEN** `lango config set knowledge.enable false` is run
- **THEN** the command SHALL fail before saving
- **AND** the error SHALL include `knowledge.enabled`
- **AND** the error SHALL include `lango config keys knowledge`

#### Scenario: Config set leaf-extension path suggests valid leaf key
- **WHEN** `lango config set agent.provider.extra openai` is run
- **THEN** the command SHALL fail before saving
- **AND** the error SHALL include `agent.provider`
- **AND** the error SHALL include `lango config keys agent`

#### Scenario: Config set creates provider map entry
- **WHEN** `lango config set providers.openai.type openai` is run on a profile with no `providers.openai` entry
- **THEN** the saved profile SHALL include `providers.openai.type` set to `openai`

#### Scenario: Config set updates map-backed struct field
- **WHEN** `lango config set providers.openai.baseUrl http://localhost:11434/v1` is run on a profile with an existing `providers.openai` entry
- **THEN** the saved profile SHALL update only the `providers.openai.baseUrl` field

#### Scenario: Config set creates nested string map value
- **WHEN** `lango config set mcp.servers.docs.env.LOG_LEVEL debug` is run on a profile with no `mcp.servers.docs` entry
- **THEN** the saved profile SHALL include `mcp.servers.docs.env.LOG_LEVEL` set to `debug`

#### Scenario: Invalid map-backed set path does not save
- **WHEN** `lango config set providers.openai.notAField value` is run
- **THEN** the command SHALL fail before saving
