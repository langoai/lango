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

## Removed Requirements

### Requirement: OAuth Provider Login (REMOVED 2026-02-14)
**Reason**: OAuth with AI providers risks account bans.
**Migration**: Use API key authentication with `${ENV_VAR}` references.

Previously, providers could be configured with `clientId`, `clientSecret`, and `scopes` for OAuth-based authentication via `lango login [provider]`. This has been removed.

## Deprecated Requirements

### Requirement: Legacy API Key Support
**Reason**: Duplication and ambiguity with `providers` map.
**Migration**: Move `agent.apiKey` to `providers.<agent.provider>.apiKey`.

#### Scenario: Legacy config detected
- **WHEN** user configuration contains `agent.apiKey`
- **THEN** system fails to start (or ignores it with a warning, depending on strictness - we choose fail for clarity)
