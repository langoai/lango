## MODIFIED Requirements

### Requirement: Workflow run command
The CLI SHALL provide `lango workflow run <file.yaml>` that parses and executes a workflow YAML file.

#### Scenario: Run without runtime reports server unavailable guidance
- **WHEN** user runs `lango workflow run code-review.flow.yaml`
- **AND** the live runtime bootstrap is unavailable
- **THEN** the CLI SHALL still display the validated workflow summary
- **AND** SHALL report that direct execution is not available from the current runtime state

#### Scenario: Run with workflow engine disabled reports guidance
- **WHEN** user runs `lango workflow run code-review.flow.yaml`
- **AND** bootstrap succeeds but the workflow engine is disabled in config
- **THEN** the CLI SHALL still display the validated workflow summary
- **AND** SHALL report that the workflow engine is not enabled
