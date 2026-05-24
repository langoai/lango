## ADDED Requirements
### Requirement: README memory completeness guard stays executable
Repository-level regressions that drop implemented `lango memory` command entries from the README quick reference SHALL be enforced by an executable test.

#### Scenario: Missing README memory entries are rejected
- **WHEN** the implemented `lango memory` CLI family remains shipped
- **THEN** an executable repository test SHALL fail if `README.md` omits those quick-reference entries
