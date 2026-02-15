## Requirements

### Requirement: Provider Registration
The system SHALL support dynamic registration of providers at startup.

#### Scenario: Register provider
- **WHEN** `Registry.Register(provider)` is called
- **THEN** the provider SHALL be stored and accessible via `Registry.Get(id)`

### Requirement: Provider Lookup
The system SHALL support looking up providers by identifier.

#### Scenario: Get existing provider
- **WHEN** `Registry.Get(id)` is called with a registered provider ID
- **THEN** it SHALL return the provider and `true`

#### Scenario: Get unknown provider
- **WHEN** `Registry.Get(id)` is called with an unknown provider ID
- **THEN** it SHALL return `nil` and `false`

### Requirement: Provider ID Normalization
The system SHALL normalize provider IDs for consistent lookup.

#### Scenario: Case insensitive lookup
- **WHEN** provider is registered as "OpenAI"
- **THEN** `Registry.Get("openai")` SHALL return the same provider

#### Scenario: Alias resolution
- **WHEN** `Registry.Get("gpt")` is called
- **THEN** it SHALL resolve to the "openai" provider if registered

#### Scenario: Claude alias
- **WHEN** `Registry.Get("claude")` is called
- **THEN** it SHALL resolve to the "anthropic" provider if registered

### Requirement: Thread-Safe Access
The system SHALL support concurrent access to the registry.

#### Scenario: Concurrent registration and lookup
- **WHEN** multiple goroutines register and lookup providers
- **THEN** the registry SHALL be thread-safe using appropriate synchronization

### Requirement: Provider Lifecycle
The system SHALL support provider initialization from configuration.

#### Scenario: Initialize from config
- **WHEN** application starts with providers in configuration
- **THEN** each configured provider SHALL be initialized and registered
- **AND** MUST return a model compatible with the ADK Model interface

### Requirement: All provider constructors accept config key as ID
All provider constructors (OpenAI, Anthropic, Gemini) SHALL accept the config map key as an explicit `id` parameter. The Supervisor SHALL pass the config key to each provider constructor so that registry lookup by config key succeeds.

#### Scenario: Gemini provider with custom config key
- **WHEN** a Gemini provider is configured with key `"gemini-api-key"` and the Supervisor initializes providers
- **THEN** the Gemini provider SHALL be registered with ID `"gemini-api-key"` and SHALL be retrievable via `registry.Get("gemini-api-key")`

#### Scenario: Anthropic provider with custom config key
- **WHEN** an Anthropic provider is configured with key `"my-anthropic"` and the Supervisor initializes providers
- **THEN** the Anthropic provider SHALL be registered with ID `"my-anthropic"` and SHALL be retrievable via `registry.Get("my-anthropic")`
