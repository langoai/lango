## MODIFIED Requirements

### Requirement: Config set uses single bootstrap with cleanup
The `config set` command SHALL bootstrap exactly once. The cfgLoader function SHALL return a cleanup function that closes the DB client. The cleanup function MUST be called via `defer` in `RunE` to ensure resources are released on all code paths (success, setConfigPath error, save error).

`config set` SHALL preserve the loaded profile's `ExplicitKeys` metadata when saving. When the set path is one of `config.ContextRelatedKeys()`, the saved explicit-key map SHALL include that path so later context auto-enable resolution treats the value as user-explicit.

Invalid dot-path errors from `config set` SHALL include actionable key discovery help. When nearby valid keys exist, the error SHALL include up to three deterministic suggestions. The command SHALL NOT save after an invalid path.

`config set` SHALL support map-backed dot paths whose map keys are dynamic user names. When a map entry is missing and the remaining path is valid for the map value type, the command SHALL create the map and entry before setting the requested value. When the final map value is a `string`, the final path segment SHALL be treated as the map key and the provided value SHALL be stored as the map value.

`config set` SHALL redact sensitive values from success output. Sensitive paths include credential-like path segments such as API keys, singular token fields, secrets, passwords, credential fields, authorization headers, PIN fields, private keys, and access keys. Redaction SHALL affect only the command output; the saved config value SHALL remain the provided value. Non-secret token-count, key-directory, key-identifier, and credential-age paths such as `agent.maxTokens`, `p2p.keyDir`, `security.signer.keyId`, and `p2p.zkp.maxCredentialAge` SHALL remain visible in success output.

`config set` SHALL support `--from-env <ENV>` as an alternative value source. When `--from-env` is present, the command SHALL require exactly one positional argument, read the config value from the named environment variable, and then follow the same set/save/output behavior as a positional value. The command SHALL treat an existing empty environment variable as a valid empty value. If the named environment variable is not set, the command SHALL fail before loading or saving config. If `--from-env` is combined with a positional value, the command SHALL fail before loading or saving config.

#### Scenario: Config set from env saves provider API key and redacts output
- **WHEN** `OPENAI_API_KEY=sk-secret lango config set providers.openai.apiKey --from-env OPENAI_API_KEY` succeeds
- **THEN** the command output SHALL include `Set providers.openai.apiKey = <redacted>`
- **AND** the command output SHALL NOT include `sk-secret`
- **AND** the saved provider API key SHALL be `sk-secret`

#### Scenario: Config set from env saves non-sensitive value and shows output
- **WHEN** `LANGO_AGENT_PROVIDER=openai lango config set agent.provider --from-env LANGO_AGENT_PROVIDER` succeeds
- **THEN** the command output SHALL include `Set agent.provider = openai`
- **AND** the saved agent provider SHALL be `openai`

#### Scenario: Config set from env accepts empty variable
- **WHEN** `EMPTY_VALUE=` is set and `lango config set providers.openai.apiKey --from-env EMPTY_VALUE` succeeds
- **THEN** the saved provider API key SHALL be an empty string
- **AND** the command output SHALL include `Set providers.openai.apiKey = <redacted>`

#### Scenario: Config set from env rejects missing variable before load
- **WHEN** `MISSING_SECRET` is unset and `lango config set providers.openai.apiKey --from-env MISSING_SECRET` is run
- **THEN** the command SHALL fail before loading config
- **AND** the error SHALL mention that environment variable `MISSING_SECRET` is not set

#### Scenario: Config set from env rejects positional value
- **WHEN** `lango config set providers.openai.apiKey raw-secret --from-env OPENAI_API_KEY` is run
- **THEN** the command SHALL fail before loading config
- **AND** the error SHALL state that `--from-env` cannot be combined with a value argument
