## MODIFIED Requirements

### Requirement: Model fields use searchable dropdown
All model selection fields in settings forms MUST use InputSearchSelect when models are fetched from API.

#### Scenario: Agent model field with fetched models
- **WHEN** FetchModelOptions returns models for the agent provider
- **THEN** model field uses InputSearchSelect type

#### Scenario: Embedding model field with filtered models
- **WHEN** embedding provider has models available
- **THEN** FetchEmbeddingModelOptions filters for embedding-pattern models
- **AND** falls back to full list if no embedding models match

#### Scenario: Esc key with open dropdown in form
- **WHEN** user presses Esc while a search-select dropdown is open in StepForm
- **THEN** editor passes Esc to form (closes dropdown) instead of exiting the form
