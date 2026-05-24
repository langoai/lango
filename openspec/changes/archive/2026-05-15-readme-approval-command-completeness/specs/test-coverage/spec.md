## ADDED Requirements
### Requirement: README approval completeness guard stays executable
Repository-level regressions that drop implemented `lango approval` command entries from the README quick reference SHALL be enforced by an executable test.

#### Scenario: Missing README approval entry is rejected
- **WHEN** the implemented `lango approval` CLI family remains shipped
- **THEN** an executable repository test SHALL fail if `README.md` omits that quick-reference entry
