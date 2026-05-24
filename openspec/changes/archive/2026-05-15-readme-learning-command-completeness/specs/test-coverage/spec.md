## ADDED Requirements
### Requirement: README learning completeness guard stays executable
Repository-level regressions that drop implemented `lango learning` command entries from the README quick reference SHALL be enforced by an executable test.

#### Scenario: Missing README learning entries are rejected
- **WHEN** the implemented `lango learning` CLI family remains shipped
- **THEN** an executable repository test SHALL fail if `README.md` omits those quick-reference entries
