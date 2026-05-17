## MODIFIED Requirements

### Requirement: Config get actionable key errors
The `config get <dot.path>` command SHALL return actionable key discovery help when the dot path cannot be resolved. When nearby valid keys exist, the error SHALL include up to three deterministic suggestions. When no nearby keys exist, the error SHALL still include a `lango config keys` discovery hint.

`config get` SHALL redact sensitive values from plain and JSON output by default. Sensitive paths SHALL use the same credential-like path matching as `config set` success output. Redaction SHALL replace sensitive leaf values with `<redacted>` and SHALL NOT mutate the loaded config object. Non-sensitive values SHALL remain visible.

`config get` SHALL support `--show-secrets` as an explicit override. When `--show-secrets` is present, the command SHALL print the raw resolved value using the selected output format.

When `config get` resolves an object, map, slice, pointer, or interface value, the command SHALL recursively redact nested sensitive leaves by their full dot path while preserving output shape. JSON output SHALL remain valid JSON after redaction.

#### Scenario: Config get redacts sensitive scalar by default
- **WHEN** `lango config get providers.openai.apiKey` reads a stored value `sk-secret`
- **THEN** the command output SHALL be `<redacted>`
- **AND** the command output SHALL NOT include `sk-secret`

#### Scenario: Config get redacts sensitive scalar JSON by default
- **WHEN** `lango config get providers.openai.apiKey --output json` reads a stored value `sk-secret`
- **THEN** the JSON output SHALL decode to `<redacted>`
- **AND** the command output SHALL NOT include `sk-secret`

#### Scenario: Config get show-secrets returns raw sensitive value
- **WHEN** `lango config get providers.openai.apiKey --show-secrets` reads a stored value `sk-secret`
- **THEN** the command output SHALL include `sk-secret`
- **AND** the command output SHALL NOT include `<redacted>`

#### Scenario: Config get redacts nested object JSON by default
- **WHEN** `lango config get providers --output json` reads a provider with `apiKey=sk-secret` and `type=openai`
- **THEN** the JSON output SHALL include the provider type
- **AND** the provider API key SHALL be `<redacted>`
- **AND** the command output SHALL NOT include `sk-secret`

#### Scenario: Config get redacts nested dynamic map values by default
- **WHEN** `lango config get mcp.servers.docs --output json` reads env `OPENAI_API_KEY=sk-secret` and env `LOG_LEVEL=debug`
- **THEN** the JSON output SHALL include `OPENAI_API_KEY` as `<redacted>`
- **AND** the JSON output SHALL include `LOG_LEVEL` as `debug`
- **AND** the command output SHALL NOT include `sk-secret`

#### Scenario: Config get redaction does not mutate loaded config
- **WHEN** `lango config get providers.openai --output json` redacts a stored provider API key
- **THEN** the in-memory loaded config still contains the raw provider API key after command execution
