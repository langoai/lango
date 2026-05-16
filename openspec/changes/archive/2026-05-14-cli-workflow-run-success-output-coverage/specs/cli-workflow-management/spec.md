## MODIFIED Requirements

### Requirement: Workflow run command
The CLI SHALL provide `lango workflow run <file.yaml>` that parses and executes a workflow YAML file.

#### Scenario: Run success writes completion details to command output
- **WHEN** user runs `lango workflow run code-review.flow.yaml`
- **AND** direct execution succeeds
- **THEN** the CLI SHALL write the execution banner, final status, and per-step output to the Cobra command output stream
