## ADDED Requirements

### Requirement: Architecture and README inventory docs include workflow validate
The public architecture inventory docs SHALL include the shipped `lango workflow validate` command instead of stopping the workflow family at history.

#### Scenario: Workflow validate stays discoverable in inventory docs
- **WHEN** a maintainer updates `docs/architecture/project-structure.md` or the README internal tree inventory
- **THEN** those inventories SHALL include the workflow `validate` surface
