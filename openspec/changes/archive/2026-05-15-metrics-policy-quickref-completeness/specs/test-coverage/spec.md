## ADDED Requirements

### Requirement: Metrics quick-reference completeness guard stays executable
Repository-level regressions that drop implemented `lango metrics` command entries from the public quick references SHALL be enforced by an executable test.

#### Scenario: Implemented metrics quick-reference entries remain listed
- **WHEN** the repository still ships the implemented `lango metrics` command family, including `lango metrics policy`
- **THEN** an executable repository test SHALL fail if `README.md` or `docs/cli/index.md` omits those quick-reference entries
