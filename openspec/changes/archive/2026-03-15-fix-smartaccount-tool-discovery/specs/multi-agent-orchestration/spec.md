## ADDED Requirements

### Requirement: Diagnostics section in orchestrator prompt
The orchestrator system prompt SHALL include a Diagnostics section instructing the orchestrator to use `builtin_list` or `builtin_health` when tools appear to be missing or a feature is not working.

#### Scenario: Orchestrator prompt contains diagnostics guidance
- **WHEN** `buildOrchestratorInstruction()` generates the orchestrator prompt
- **THEN** the prompt SHALL contain a "Diagnostics" section
- **AND** the section SHALL reference `builtin_health` as the diagnostic tool
