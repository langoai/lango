## MODIFIED Requirements

### Requirement: Config set uses single bootstrap with cleanup
The `config set` command SHALL bootstrap exactly once. The cfgLoader function SHALL return a cleanup function that closes the DB client. The cleanup function MUST be called via `defer` in `RunE` to ensure resources are released on all code paths (success, setConfigPath error, save error).

`config set` SHALL preserve the loaded profile's `ExplicitKeys` metadata when saving. When the set path is one of `config.ContextRelatedKeys()`, the saved explicit-key map SHALL include that path so later context auto-enable resolution treats the value as user-explicit.

Invalid dot-path errors from `config set` SHALL include actionable key discovery help. When nearby valid keys exist, the error SHALL include up to three deterministic suggestions. The command SHALL NOT save after an invalid path.

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

### Requirement: Config get actionable key errors
The `config get <dot.path>` command SHALL return actionable key discovery help when the dot path cannot be resolved. When nearby valid keys exist, the error SHALL include up to three deterministic suggestions. When no nearby keys exist, the error SHALL still include a `lango config keys` discovery hint.

#### Scenario: Config get suggests nearby key on invalid path
- **WHEN** `lango config get agent.providr` is run
- **THEN** the command SHALL fail with an error containing `agent.provider`
- **AND** the error SHALL include `lango config keys agent`

#### Scenario: Config get unknown top-level path gives discovery hint
- **WHEN** `lango config get made.up.path` is run
- **THEN** the command SHALL fail with an error containing `lango config keys`

#### Scenario: Config get leaf-extension path suggests valid leaf key
- **WHEN** `lango config get agent.provider.extra` is run
- **THEN** the command SHALL fail with an error containing `agent.provider`
- **AND** the error SHALL include `lango config keys agent`
