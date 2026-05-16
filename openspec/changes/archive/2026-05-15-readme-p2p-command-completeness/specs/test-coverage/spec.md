## ADDED Requirements
### Requirement: README P2P completeness guard stays executable
Repository-level regressions that drop implemented `workspace`, `team`, or `zkp` P2P command entries from the README quick reference SHALL be enforced by an executable test.

#### Scenario: Missing README P2P family entries are rejected
- **WHEN** the implemented `workspace`, `team`, and `zkp` P2P command families remain shipped
- **THEN** an executable repository test SHALL fail if `README.md` omits those quick-reference entries
