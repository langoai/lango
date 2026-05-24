## ADDED Requirements
### Requirement: README run completeness guard stays executable
Repository-level regressions that drop implemented `lango run` command entries from the README quick reference SHALL be enforced by an executable test.

#### Scenario: Missing README run entries are rejected
- **WHEN** the implemented `lango run` CLI family remains shipped
- **THEN** an executable repository test SHALL fail if `README.md` omits those quick-reference entries
