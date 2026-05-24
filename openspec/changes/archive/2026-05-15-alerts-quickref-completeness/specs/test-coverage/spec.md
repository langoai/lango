## ADDED Requirements

### Requirement: Alerts quick-reference completeness guard stays executable
Repository-level regressions that drop implemented `lango alerts` command entries from the public quick references SHALL be enforced by an executable test.

#### Scenario: Implemented alerts quick-reference entries remain listed
- **WHEN** the repository still ships the implemented `lango alerts list` and `lango alerts summary` commands
- **THEN** an executable repository test SHALL fail if `README.md` or `docs/cli/index.md` omits those quick-reference entries
