## ADDED Requirements

### Requirement: Architecture project-structure docs stay aligned with the current graph and metrics CLI surface
The public architecture project-structure reference SHALL list the currently implemented `cli/graph/` and `cli/metrics/` command families rather than outdated subsets.

#### Scenario: Project-structure graph and metrics rows stay truthful
- **WHEN** a maintainer updates `docs/architecture/project-structure.md`
- **THEN** the `cli/graph/` row SHALL include `add`, `export`, and `import`
- **AND** the `cli/metrics/` row SHALL include `policy`
