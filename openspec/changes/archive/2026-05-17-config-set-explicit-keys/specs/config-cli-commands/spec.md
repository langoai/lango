## MODIFIED Requirements

### Requirement: Config set uses single bootstrap with cleanup
The `config set` command SHALL bootstrap exactly once. The cfgLoader function SHALL return a cleanup function that closes the DB client. The cleanup function MUST be called via `defer` in `RunE` to ensure resources are released on all code paths (success, setConfigPath error, save error).

`config set` SHALL preserve the loaded profile's `ExplicitKeys` metadata when saving. When the set path is one of `config.ContextRelatedKeys()`, the saved explicit-key map SHALL include that path so later context auto-enable resolution treats the value as user-explicit.

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
