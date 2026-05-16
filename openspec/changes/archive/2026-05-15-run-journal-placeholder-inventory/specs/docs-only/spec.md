## ADDED Requirements

### Requirement: Run inventory docs keep the `journal <run-id>` placeholder
The architecture inventory and README internal tree SHALL describe the run journal command with its current placeholder.

#### Scenario: Placeholder stays visible
- **WHEN** a maintainer updates `docs/architecture/project-structure.md` or the README internal tree inventory
- **THEN** they SHALL describe `journal <run-id>` instead of a bare `journal`
