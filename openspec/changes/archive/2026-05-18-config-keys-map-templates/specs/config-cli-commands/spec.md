## MODIFIED Requirements

### Requirement: Config set uses single bootstrap with cleanup
The `config set` command SHALL bootstrap exactly once. The cfgLoader function SHALL return a cleanup function that closes the DB client. The cleanup function MUST be called via `defer` in `RunE` to ensure resources are released on all code paths (success, setConfigPath error, save error).

`config set` SHALL preserve the loaded profile's `ExplicitKeys` metadata when saving. When the set path is one of `config.ContextRelatedKeys()`, the saved explicit-key map SHALL include that path so later context auto-enable resolution treats the value as user-explicit.

Invalid dot-path errors from `config set` SHALL include actionable key discovery help. When nearby valid keys exist, the error SHALL include up to three deterministic suggestions. The command SHALL NOT save after an invalid path.

`config set` SHALL support map-backed dot paths whose map keys are dynamic user names. When a map entry is missing and the remaining path is valid for the map value type, the command SHALL create the map and entry before setting the requested value. When the final map value is a `string`, the final path segment SHALL be treated as the map key and the provided value SHALL be stored as the map value.

`config keys` SHALL list runnable templates for string-keyed map-backed paths that `config set` accepts. Struct-valued maps SHALL use a `<name>` placeholder for the dynamic map key and list the struct leaf fields. String-valued maps SHALL use a `<key>` placeholder for the dynamic leaf key. Dynamic map templates SHALL omit fields whose leaf type is not accepted by `config set`, such as `time.Duration` fields.

`config set` SHALL redact sensitive values from success output. Sensitive paths include credential-like path segments such as API keys, singular token fields, secrets, passwords, credential fields, authorization headers, PIN fields, private keys, and access keys. Redaction SHALL affect only the command output; the saved config value SHALL remain the provided value. Non-secret token-count, key-directory, key-identifier, and credential-age paths such as `agent.maxTokens`, `p2p.keyDir`, `security.signer.keyId`, and `p2p.zkp.maxCredentialAge` SHALL remain visible in success output.

`config set` SHALL support `--from-env <ENV>` as an alternative value source. When `--from-env` is present, the command SHALL require exactly one positional argument, read the config value from the named environment variable, and then follow the same set/save/output behavior as a positional value. The command SHALL treat an existing empty environment variable as a valid empty value. If the named environment variable is not set, the command SHALL fail before loading or saving config. If `--from-env` is combined with a positional value, the command SHALL fail before loading or saving config.

#### Scenario: Config keys lists provider map templates
- **WHEN** `lango config keys providers` is run
- **THEN** the command SHALL include `providers.<name>.type`
- **AND** the command SHALL include `providers.<name>.apiKey`
- **AND** the command SHALL include `providers.<name>.baseUrl`

#### Scenario: Config keys lists nested string map templates
- **WHEN** `lango config keys mcp.servers` is run
- **THEN** the command SHALL include `mcp.servers.<name>.env.<key>`
- **AND** the command SHALL include `mcp.servers.<name>.headers.<key>`
- **AND** the command SHALL NOT include `mcp.servers.<name>.timeout`

#### Scenario: Config keys lists auth provider map templates
- **WHEN** `lango config keys auth.providers` is run
- **THEN** the command SHALL include `auth.providers.<name>.clientSecret`
