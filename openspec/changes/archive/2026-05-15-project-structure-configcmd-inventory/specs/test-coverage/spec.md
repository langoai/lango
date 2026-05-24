## ADDED Requirements

### Requirement: Architecture config CLI truthfulness guard stays executable
Repository-level regressions that let the architecture project-structure docs omit the shipped `cli/configcmd/` package or its current command surface SHALL be enforced by an executable test.

#### Scenario: Project-structure config row remains truthful
- **WHEN** the repository still ships the `lango config` management surface with `get`, `set`, and `keys`
- **THEN** an executable repository test SHALL fail if `docs/architecture/project-structure.md` omits `cli/configcmd/` or stops describing those current command surfaces
