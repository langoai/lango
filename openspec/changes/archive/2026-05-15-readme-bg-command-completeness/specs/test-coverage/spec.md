## ADDED Requirements
### Requirement: README background completeness guard stays executable
Repository-level regressions that drop implemented `lango bg` command entries from the README quick reference SHALL be enforced by an executable test.

#### Scenario: Missing README background entries are rejected
- **WHEN** the implemented `lango bg` CLI family remains shipped
- **THEN** an executable repository test SHALL fail if `README.md` omits those quick-reference entries
