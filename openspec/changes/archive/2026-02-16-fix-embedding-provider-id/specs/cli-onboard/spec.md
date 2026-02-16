## MODIFIED Requirements

### Requirement: Embedding form provider selection

The onboard TUI embedding form SHALL display the user's registered provider IDs from the providers map plus `"local"` as options, instead of hardcoded type strings. When a provider ID is selected, the form SHALL set `ProviderID` on the embedding config and auto-resolve the `Provider` type field. When `"local"` is selected, `ProviderID` SHALL be cleared and `Provider` SHALL be set to `"local"`.

#### Scenario: Provider options from registered providers
- **WHEN** the user has providers `"gemini-1"` and `"my-openai"` registered
- **THEN** the embedding provider dropdown SHALL show `["gemini-1", "local", "my-openai"]` (sorted, with "local" always included)

#### Scenario: Selecting a registered provider
- **WHEN** the user selects `"my-openai"` (type: openai) from the embedding provider dropdown
- **THEN** `embedding.providerID` SHALL be set to `"my-openai"` and `embedding.provider` SHALL be auto-resolved to `"openai"`

#### Scenario: Selecting local provider
- **WHEN** the user selects `"local"` from the embedding provider dropdown
- **THEN** `embedding.providerID` SHALL be empty and `embedding.provider` SHALL be `"local"`

#### Scenario: Current value display
- **WHEN** `embedding.providerID` is set to `"gemini-1"`
- **THEN** the form SHALL show `"gemini-1"` as the current selected value
