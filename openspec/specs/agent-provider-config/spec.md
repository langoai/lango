## Purpose

Capability spec for agent-provider-config. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: Provider Configuration
The system SHALL allow configuring AI providers with an API key. All provider credentials SHALL be configured in the `providers` map.

#### Scenario: Provider with API Key
- **WHEN** `lango.json` includes a provider with `apiKey`
- **THEN** system initializes the provider using the API key directly

#### Scenario: Provider with environment variable reference
- **WHEN** `apiKey` contains `${ENV_VAR}` pattern
- **THEN** system expands the environment variable before use
- **AND** `lango doctor` reports the key as securely configured

#### Scenario: Provider with plaintext API key
- **WHEN** `apiKey` contains a literal key (not an `${ENV_VAR}` reference)
- **THEN** system initializes the provider using the key
- **AND** `lango doctor` warns the user to use environment variables or encrypted profiles

#### Scenario: Provider without credentials
- **WHEN** `apiKey` is empty or missing
- **THEN** system logs a warning during initialization

#### Scenario: Missing provider configuration
- **WHEN** `agent.provider` is set to "google"
- **BUT** `providers.google` is missing or empty
- **THEN** system fails to start with a configuration error

### Requirement: Agent configuration supports prompts directory
The `AgentConfig` struct SHALL include a `PromptsDir` field (mapstructure: "promptsDir") specifying the directory containing section `.md` files. The system SHALL support three-tier precedence: PromptsDir > SystemPromptPath > built-in defaults.

#### Scenario: PromptsDir configured
- **WHEN** AgentConfig.PromptsDir is set to a valid directory path
- **THEN** the system SHALL load prompt sections from .md files in that directory

#### Scenario: Legacy SystemPromptPath only
- **WHEN** AgentConfig.PromptsDir is empty but SystemPromptPath is set
- **THEN** the file content SHALL replace the Identity section only, and all other default sections SHALL remain

#### Scenario: No prompt configuration
- **WHEN** both PromptsDir and SystemPromptPath are empty
- **THEN** the system SHALL use the built-in default sections including conversation rules

### Requirement: OAuth provider login is removed
The system SHALL NOT support OAuth-based AI provider login fields such as `clientId`, `clientSecret`, or `scopes` in provider configuration.

#### Scenario: OAuth-style provider login fields are rejected
- **WHEN** provider configuration includes OAuth-style provider login fields
- **THEN** the system SHALL reject or ignore those fields
- **AND** provider authentication SHALL remain API-key based

### Requirement: Legacy API key field is removed
The system SHALL NOT use `agent.apiKey` as the canonical provider credential source. Provider credentials SHALL live under `providers.<agent.provider>.apiKey`.

#### Scenario: Legacy config detected
- **WHEN** user configuration contains `agent.apiKey`
- **THEN** the system SHALL fail startup or emit a migration warning directing the user to `providers.<agent.provider>.apiKey`
