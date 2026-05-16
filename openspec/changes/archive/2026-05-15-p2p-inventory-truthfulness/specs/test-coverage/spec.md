## ADDED Requirements

### Requirement: P2P inventory truthfulness guard stays executable
Repository-level regressions that let architecture or README inventory docs describe stale P2P CLI subsets SHALL be enforced by an executable test.

#### Scenario: P2P inventory remains truthful
- **WHEN** the repository still ships P2P workspace, git, provenance, team, and ZKP command families
- **THEN** an executable repository test SHALL fail if `docs/architecture/project-structure.md` or `README.md` stops describing those current command surfaces
