## ADDED Requirements

### Requirement: Architecture project-structure docs include the current config CLI surface
The public architecture project-structure reference SHALL include the shipped `cli/configcmd/` package and its current configuration-management command surface.

#### Scenario: Project-structure config row stays truthful
- **WHEN** a maintainer updates `docs/architecture/project-structure.md`
- **THEN** it SHALL include a `cli/configcmd/` row
- **AND** that row SHALL describe `lango config list`, `create`, `use`, `delete`, `import`, `export`, `get`, `set`, `keys`, and `validate`
