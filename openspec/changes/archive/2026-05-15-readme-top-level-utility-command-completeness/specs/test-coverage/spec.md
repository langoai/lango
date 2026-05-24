## ADDED Requirements

### Requirement: README top-level utility completeness guard stays executable
Repository-level regressions that drop implemented `lango version` and `lango health` command entries from the README quick reference SHALL be enforced by an executable test.

#### Scenario: Implemented top-level utility quick-reference entries remain listed
- **WHEN** the repository still ships the implemented `lango version` and `lango health` commands
- **THEN** an executable repository test SHALL fail if `README.md` omits those quick-reference entries
