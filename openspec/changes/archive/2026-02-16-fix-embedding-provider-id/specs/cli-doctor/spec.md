## MODIFIED Requirements

### Requirement: Embedding doctor check uses unified resolver

The embedding doctor check SHALL use `Config.ResolveEmbeddingProvider()` for validation instead of hardcoded provider type switch statements and name-based API key lookups.

#### Scenario: ProviderID resolves successfully
- **WHEN** `embedding.providerID` is set to a valid provider with a supported type and API key
- **THEN** the check SHALL pass

#### Scenario: ProviderID not found
- **WHEN** `embedding.providerID` is set to a non-existent provider ID
- **THEN** the check SHALL fail with a message indicating the provider ID was not found

#### Scenario: Cloud provider with no API key
- **WHEN** `embedding.providerID` references a cloud provider (non-local) with an empty API key
- **THEN** the check SHALL fail with a message indicating no API key is configured

#### Scenario: Local provider needs no API key
- **WHEN** `embedding.provider` is `"local"`
- **THEN** the check SHALL pass without requiring an API key

#### Scenario: Neither provider configured
- **WHEN** both `embedding.providerID` and `embedding.provider` are empty
- **THEN** the check SHALL skip with "not configured" message
