## ADDED Requirements
### Requirement: README sandbox completeness guard stays executable
Repository-level regressions that drop implemented `lango sandbox` command entries from the README quick reference SHALL be enforced by an executable test.

#### Scenario: Missing README sandbox entries are rejected
- **WHEN** the implemented `lango sandbox` CLI family remains shipped
- **THEN** an executable repository test SHALL fail if `README.md` omits those quick-reference entries
