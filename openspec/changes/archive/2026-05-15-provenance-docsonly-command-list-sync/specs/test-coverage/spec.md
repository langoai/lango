## ADDED Requirements

### Requirement: README provenance completeness guard stays executable
Repository-level regressions that drop implemented `lango provenance` command entries from the README quick reference SHALL be enforced by an executable test.

#### Scenario: Implemented provenance quick-reference entries remain listed
- **WHEN** the repository still ships the implemented `lango provenance` command family
- **THEN** an executable repository test SHALL fail if `README.md` or the main `docs-only` spec omits those quick-reference entries
