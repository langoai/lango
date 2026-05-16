## ADDED Requirements

### Requirement: Onboard test step shares the workbench readiness contract

The onboard Test Configuration step SHALL evaluate provider/model/API-key completeness with the same agent-readiness rules used by the workbench setup-recovery flow.

#### Scenario: Missing remote API key fails onboard validation
- **WHEN** the selected agent provider and model target a non-ollama provider and the provider API key is empty
- **THEN** the Test Configuration step SHALL fail the API key check
- **AND** SHALL continue treating the profile as incomplete until the key is provided

#### Scenario: Ollama passes onboard validation without an API key
- **WHEN** the selected agent provider and model target an ollama provider and the provider API key is empty
- **THEN** the Test Configuration step SHALL pass the API key check
- **AND** SHALL keep that profile eligible for ready-state workbench behavior after save
