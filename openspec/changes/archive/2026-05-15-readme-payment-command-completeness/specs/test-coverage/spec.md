## ADDED Requirements
### Requirement: README payment completeness guard stays executable
Repository-level regressions that drop implemented `lango payment` command entries from the README quick reference SHALL be enforced by an executable test.

#### Scenario: Missing README payment entries are rejected
- **WHEN** the implemented `lango payment` CLI family remains shipped
- **THEN** an executable repository test SHALL fail if `README.md` omits those quick-reference entries
