## ADDED Requirements
### Requirement: README smart-account completeness guard stays executable
Repository-level regressions that drop implemented `lango account` command entries from the README quick reference SHALL be enforced by an executable test.

#### Scenario: Missing README smart-account entries are rejected
- **WHEN** the implemented `lango account` CLI family remains shipped
- **THEN** an executable repository test SHALL fail if `README.md` omits those quick-reference entries
