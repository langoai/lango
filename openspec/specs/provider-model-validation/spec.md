## Purpose

Heuristic prefix-based blocklist to detect cross-provider model routing errors at validation time and runtime.

## Requirements

### Requirement: Heuristic model-provider compatibility validation
The system SHALL provide a `ValidateModelProvider(providerType, model string) error` function that detects obviously incompatible model-provider combinations using prefix-based blocklists.

#### Scenario: Empty model passes validation
- **WHEN** `ValidateModelProvider` is called with any provider type and an empty model string
- **THEN** it SHALL return nil

#### Scenario: Correct model-provider pair passes
- **WHEN** `ValidateModelProvider("gemini", "gemini-3-flash-preview")` is called
- **THEN** it SHALL return nil

#### Scenario: Cross-provider model detected
- **WHEN** `ValidateModelProvider("gemini", "gpt-5.3-codex")` is called
- **THEN** it SHALL return an error wrapping `ErrModelProviderMismatch`

#### Scenario: Case-insensitive matching
- **WHEN** `ValidateModelProvider("gemini", "GPT-5.3-codex")` is called
- **THEN** it SHALL return an error (case-insensitive prefix match)

#### Scenario: Ollama and GitHub have no exclusions
- **WHEN** `ValidateModelProvider("ollama", "gpt-4o")` is called
- **THEN** it SHALL return nil (ollama can host any model)

#### Scenario: Unknown provider type has no exclusions
- **WHEN** `ValidateModelProvider("custom", "any-model")` is called
- **THEN** it SHALL return nil
