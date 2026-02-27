## MODIFIED Requirements

### Requirement: Model Fetcher API
The settings package SHALL export `FetchModelOptions` and `NewProviderFromConfig` as public functions so other CLI packages (e.g., onboard) can reuse model auto-fetch logic.

#### Scenario: Exported function availability
- **WHEN** another package imports the settings package
- **THEN** `settings.FetchModelOptions(providerID, cfg, currentModel)` SHALL be callable
- **AND** `settings.NewProviderFromConfig(id, pCfg)` SHALL be callable

### Requirement: Configuration Coverage
The settings editor SHALL support editing all configuration sections. The `NewProviderForm` type options SHALL include "github" alongside openai, anthropic, gemini, and ollama.

#### Scenario: Provider form includes github
- **WHEN** user opens the provider add/edit form
- **THEN** the Type select field options SHALL include "github"
