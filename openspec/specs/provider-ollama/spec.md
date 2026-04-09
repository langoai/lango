## Purpose

Capability spec for provider-ollama. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: Configuration
The system SHALL support "ollama" as a distinct provider type in configuration.

#### Scenario: Default Configuration
- **WHEN** a provider is configured with type "ollama"
- **AND** no `baseUrl` is specified
- **THEN** it SHALL use the default Ollama base URL `http://localhost:11434/v1`
- **AND** it SHALL function as an OpenAI-compatible provider

#### Scenario: Custom Configuration
- **WHEN** a provider is configured with type "ollama"
- **AND** a custom `baseUrl` is provided (e.g., `http://remote-host:11434/v1`)
- **THEN** it SHALL use the provided base URL
- **AND** it SHALL function as an OpenAI-compatible provider

### Requirement: Onboarding
The system SHALL support Ollama during the initial setup wizard.

#### Scenario: Provider Selection
- **WHEN** the user runs `lango onboard`
- **THEN** "Ollama" SHALL be listed as an available AI provider

#### Scenario: Simplified Setup
- **WHEN** the user selects "Ollama" in the wizard
- **THEN** the wizard SHALL NOT request an API key
- **AND** the wizard SHALL configure the provider with type "ollama" in `lango.json`
