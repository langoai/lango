## ADDED Requirements
### Requirement: README metrics completeness guard stays executable
Repository-level regressions that drop implemented `lango metrics` command entries from the README quick reference SHALL be enforced by an executable test.

#### Scenario: Missing README metrics entries are rejected
- **WHEN** the implemented `lango metrics` CLI family remains shipped
- **THEN** an executable repository test SHALL fail if `README.md` omits those quick-reference entries
