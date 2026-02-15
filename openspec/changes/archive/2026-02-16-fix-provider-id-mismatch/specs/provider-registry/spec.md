## MODIFIED Requirements

### Requirement: All provider constructors accept config key as ID
All provider constructors (OpenAI, Anthropic, Gemini) SHALL accept the config map key as an explicit `id` parameter. The Supervisor SHALL pass the config key to each provider constructor so that registry lookup by config key succeeds.

#### Scenario: Gemini provider with custom config key
- **WHEN** a Gemini provider is configured with key `"gemini-api-key"` and the Supervisor initializes providers
- **THEN** the Gemini provider SHALL be registered with ID `"gemini-api-key"` and SHALL be retrievable via `registry.Get("gemini-api-key")`

#### Scenario: Anthropic provider with custom config key
- **WHEN** an Anthropic provider is configured with key `"my-anthropic"` and the Supervisor initializes providers
- **THEN** the Anthropic provider SHALL be registered with ID `"my-anthropic"` and SHALL be retrievable via `registry.Get("my-anthropic")`
