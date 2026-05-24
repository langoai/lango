## ADDED Requirements

### Requirement: Workflow inventory docs keep the `validate <file>` placeholder
The architecture inventory and README internal tree SHALL describe the workflow validate command with its current placeholder.

#### Scenario: Placeholder stays visible
- **WHEN** a maintainer updates `docs/architecture/project-structure.md` or the README internal tree inventory
- **THEN** they SHALL describe `validate <file>` instead of a bare `validate`
