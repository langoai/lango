## MODIFIED Requirements

### Requirement: Anthropic provider constructor accepts explicit ID
The Anthropic provider constructor SHALL accept an `id` string parameter and use it as the provider's registry identity, instead of hardcoding `"anthropic"`.

#### Scenario: Custom ID registration
- **WHEN** `NewProvider("my-claude", "sk-ant-xxx")` is called
- **THEN** the returned provider's `ID()` method SHALL return `"my-claude"`

#### Scenario: Default ID registration
- **WHEN** `NewProvider("anthropic", "sk-ant-xxx")` is called
- **THEN** the returned provider's `ID()` method SHALL return `"anthropic"`
