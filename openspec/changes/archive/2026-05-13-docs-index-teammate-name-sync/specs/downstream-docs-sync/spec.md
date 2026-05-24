## ADDED Requirements

### Requirement: Home page uses current built-in teammate names
The public home page SHALL describe multi-agent orchestration using current built-in teammate names rather than legacy examples.

#### Scenario: Home page feature card uses current built-in teammates
- **WHEN** a user reads the multi-agent feature card in `docs/index.md`
- **THEN** the card SHALL use current built-in teammate names such as `Operator`, `Librarian`, `Planner`, or `Vault`
- **AND** it SHALL NOT describe the system with legacy built-in names such as `Executor`, `Researcher`, or `Memory Manager`
