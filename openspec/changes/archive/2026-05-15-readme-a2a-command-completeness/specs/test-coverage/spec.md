## ADDED Requirements
### Requirement: README A2A completeness guard stays executable
Repository-level regressions that drop implemented `lango a2a` command entries from the README quick reference SHALL be enforced by an executable test.

#### Scenario: Missing README A2A entries are rejected
- **WHEN** the implemented `lango a2a` CLI family remains shipped
- **THEN** an executable repository test SHALL fail if `README.md` omits those quick-reference entries
