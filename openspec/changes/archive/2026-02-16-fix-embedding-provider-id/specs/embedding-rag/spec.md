## ADDED Requirements

### Requirement: ProviderID-based embedding provider resolution

The `EmbeddingConfig` SHALL support a `ProviderID` field that references a key in the `Config.Providers` map. When `ProviderID` is set, the embedding backend type and API key SHALL be resolved from the referenced provider's `Type` and `APIKey` fields using the `ProviderTypeToEmbeddingType` mapping.

#### Scenario: ProviderID resolves gemini provider
- **WHEN** `embedding.providerID` is set to `"gemini-1"` and `providers["gemini-1"]` has type `"gemini"` and a valid API key
- **THEN** the embedding backend type SHALL be `"google"` and the API key SHALL be the provider's API key

#### Scenario: ProviderID resolves openai provider
- **WHEN** `embedding.providerID` is set to `"my-openai"` and `providers["my-openai"]` has type `"openai"` and a valid API key
- **THEN** the embedding backend type SHALL be `"openai"` and the API key SHALL be the provider's API key

#### Scenario: ProviderID resolves ollama provider
- **WHEN** `embedding.providerID` is set to `"my-ollama"` and `providers["my-ollama"]` has type `"ollama"`
- **THEN** the embedding backend type SHALL be `"local"` and no API key is required

#### Scenario: ProviderID references unsupported type
- **WHEN** `embedding.providerID` references a provider with type `"anthropic"` (no embedding support)
- **THEN** the resolver SHALL return empty backend type and empty API key

#### Scenario: ProviderID not found in providers map
- **WHEN** `embedding.providerID` is set to an ID that does not exist in the providers map
- **THEN** the resolver SHALL return empty backend type and empty API key

### Requirement: ProviderID takes precedence over legacy Provider

When both `ProviderID` and `Provider` fields are set, the system SHALL use `ProviderID` for resolution and ignore the `Provider` field.

#### Scenario: Both fields set
- **WHEN** `embedding.providerID` is `"gemini-1"` and `embedding.provider` is `"openai"`
- **THEN** the resolver SHALL use `"gemini-1"` and return the gemini provider's type and key

### Requirement: Legacy Provider field backward compatibility

When `ProviderID` is empty, the system SHALL fall back to the existing `Provider` field behavior, resolving API keys by searching the providers map for matching types.

#### Scenario: Legacy local provider
- **WHEN** `embedding.providerID` is empty and `embedding.provider` is `"local"`
- **THEN** the backend type SHALL be `"local"` with no API key required

#### Scenario: Legacy openai provider
- **WHEN** `embedding.providerID` is empty and `embedding.provider` is `"openai"` and a provider with type `"openai"` exists
- **THEN** the backend type SHALL be `"openai"` and the API key SHALL be resolved from the matching provider

#### Scenario: Neither field configured
- **WHEN** both `embedding.providerID` and `embedding.provider` are empty
- **THEN** the embedding system SHALL be disabled
