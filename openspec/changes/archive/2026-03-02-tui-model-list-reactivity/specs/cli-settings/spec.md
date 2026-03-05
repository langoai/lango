## MODIFIED Requirements

### Requirement: Agent form reactive model list
The Agent configuration form SHALL wire `OnChange` on the provider field to asynchronously fetch and update the model field's options when the provider changes.

#### Scenario: Provider change triggers model refresh
- **WHEN** a user changes the provider field in the Agent form
- **THEN** the model field SHALL show a loading indicator and asynchronously fetch models from the new provider

#### Scenario: Fallback provider change triggers fallback model refresh
- **WHEN** a user changes the fallback provider field in the Agent form
- **THEN** the fallback model field SHALL asynchronously refresh its options from the new fallback provider

### Requirement: Knowledge forms reactive model list
The Observational Memory, Embedding, and Librarian configuration forms SHALL wire `OnChange` on their provider fields to refresh the corresponding model field options.

#### Scenario: OM provider change triggers OM model refresh
- **WHEN** a user changes the OM provider field
- **THEN** the OM model field SHALL asynchronously fetch models from the new provider (or agent provider if empty)

#### Scenario: Embedding provider change triggers embedding model refresh
- **WHEN** a user changes the embedding provider field
- **THEN** the embedding model field SHALL asynchronously fetch embedding-filtered models

#### Scenario: Librarian provider change triggers librarian model refresh
- **WHEN** a user changes the librarian provider field
- **THEN** the librarian model field SHALL asynchronously fetch models from the new provider (or agent provider if empty)

### Requirement: Async Cmd wrappers for model fetching
The settings package SHALL provide `FetchModelOptionsCmd()` and `FetchEmbeddingModelOptionsCmd()` functions that return `tea.Cmd` for async model fetching, producing `FieldOptionsLoadedMsg` results.

#### Scenario: FetchModelOptionsCmd returns loaded message
- **WHEN** `FetchModelOptionsCmd("model", "openai", cfg, "")` is executed
- **THEN** it SHALL return a `FieldOptionsLoadedMsg` with `FieldKey="model"` and `ProviderID="openai"`
