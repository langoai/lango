## ADDED Requirements

### Requirement: Workflow CLI uses workflow state store capability
Workflow CLI commands MUST obtain workflow state persistence through a storage facade capability instead of constructing a state store from a generic Ent client in the CLI layer.

#### Scenario: Workflow engine initialization uses facade state store
- **WHEN** the workflow CLI initializes a workflow engine
- **THEN** it resolves the workflow state store from the storage facade capability
