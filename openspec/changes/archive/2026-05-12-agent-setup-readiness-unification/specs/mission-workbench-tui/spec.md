## ADDED Requirements

### Requirement: Workbench readiness signals share one profile-readiness contract

The standalone workbench SHALL derive its setup-required header, empty-state guidance, and composer placeholder from one shared agent-readiness evaluation.

#### Scenario: Missing remote API key keeps all workbench recovery cues aligned
- **WHEN** bare `lango` renders with a non-ollama provider ID and model configured but the provider API key is empty
- **THEN** the header SHALL show `Model: Setup required`
- **AND** the empty-state body SHALL continue setup guidance
- **AND** the empty composer placeholder SHALL continue setup-first guidance

#### Scenario: Ollama remains ready without an API key
- **WHEN** bare `lango` renders with an ollama provider ID and model configured but no API key
- **THEN** the header SHALL show the configured provider and model summary
- **AND** the empty-state body SHALL omit setup guidance
- **AND** the empty composer placeholder SHALL show starter prompts
