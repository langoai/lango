## MODIFIED Requirements

### Requirement: Embedding provider resolution
The system SHALL resolve the embedding backend via two paths:
1. `ProviderID` — looks up the provider in the providers map and resolves backend type and API key.
2. `Provider = "local"` — uses local (Ollama) embeddings with no API key.

If neither `ProviderID` nor `Provider = "local"` is set, the embedding system SHALL be disabled.

#### Scenario: ProviderID resolves from providers map
- **WHEN** `embedding.providerID` is set to a valid key in the providers map
- **THEN** the backend type and API key SHALL be resolved from that provider entry

#### Scenario: Local provider needs no API key
- **WHEN** `embedding.provider` is set to `"local"`
- **THEN** the backend type SHALL be `"local"` with no API key

#### Scenario: Neither configured
- **WHEN** both `embedding.providerID` and `embedding.provider` are empty
- **THEN** the embedding system SHALL be disabled

## REMOVED Requirements

### Requirement: Legacy provider type-search fallback
**Reason**: The fallback that searched the providers map by type string (e.g., `provider: "openai"` scanning for any provider with `type: "openai"`) is removed. Users must use `providerID` to reference specific provider entries.
**Migration**: Set `embedding.providerID` to the key of your provider in the providers map instead of `embedding.provider`.
