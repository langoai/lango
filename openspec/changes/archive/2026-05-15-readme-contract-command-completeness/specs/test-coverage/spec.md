## ADDED Requirements
### Requirement: README contract completeness guard stays executable
Repository-level regressions that drop implemented `lango contract` command entries from the README quick reference SHALL be enforced by an executable test.

#### Scenario: Missing README contract entries are rejected
- **WHEN** the implemented `lango contract` CLI family remains shipped
- **THEN** an executable repository test SHALL fail if `README.md` omits those quick-reference entries
