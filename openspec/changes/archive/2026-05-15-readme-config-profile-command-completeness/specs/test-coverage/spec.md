## ADDED Requirements
### Requirement: README config profile-management completeness guard stays executable
Repository-level regressions that drop implemented `lango config` profile-management command entries from the README quick reference SHALL be enforced by an executable test.

#### Scenario: Missing README config profile entries are rejected
- **WHEN** the implemented `lango config` profile-management CLI family remains shipped
- **THEN** an executable repository test SHALL fail if `README.md` omits those quick-reference entries
