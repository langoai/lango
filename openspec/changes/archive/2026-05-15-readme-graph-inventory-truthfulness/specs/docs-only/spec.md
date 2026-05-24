## ADDED Requirements

### Requirement: README internal tree stays aligned with the current graph CLI surface
The README internal tree inventory SHALL include the currently implemented graph command family rather than an outdated subset.

#### Scenario: README graph inventory stays truthful
- **WHEN** a maintainer updates the README internal tree inventory
- **THEN** the `graph/` row SHALL include `add`, `export`, and `import`
