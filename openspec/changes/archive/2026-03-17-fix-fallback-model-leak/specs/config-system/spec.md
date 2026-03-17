## ADDED Requirements

### Requirement: Fallback provider existence validation
`config.Validate()` SHALL verify that `agent.fallbackProvider` (when set) references an existing key in the `providers` map.

#### Scenario: Fallback provider not in providers map
- **WHEN** `agent.fallbackProvider` is set to a value not present in the `providers` map
- **THEN** validation SHALL fail with an error identifying the missing provider

#### Scenario: Fallback provider exists
- **WHEN** `agent.fallbackProvider` references a valid key in the `providers` map
- **THEN** validation SHALL pass (no error for this check)

### Requirement: Provider-model compatibility validation at startup
`config.Validate()` SHALL check both primary (`agent.provider`/`agent.model`) and fallback (`agent.fallbackProvider`/`agent.fallbackModel`) pairs for model-provider compatibility using `ValidateModelProvider`.

#### Scenario: Primary model incompatible with provider type
- **WHEN** `agent.model` is `gpt-5.3-codex` and `agent.provider` references a gemini-type provider
- **THEN** validation SHALL fail with an error describing the mismatch

#### Scenario: Fallback model incompatible with fallback provider type
- **WHEN** `agent.fallbackModel` is `claude-sonnet-4-5-20250514` and `agent.fallbackProvider` references an openai-type provider
- **THEN** validation SHALL fail with an error describing the mismatch
