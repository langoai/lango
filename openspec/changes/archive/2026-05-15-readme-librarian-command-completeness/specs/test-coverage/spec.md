## ADDED Requirements
### Requirement: README librarian completeness guard stays executable
Repository-level regressions that drop implemented `lango librarian` command entries from the README quick reference SHALL be enforced by an executable test.

#### Scenario: Missing README librarian entries are rejected
- **WHEN** the implemented `lango librarian` CLI family remains shipped
- **THEN** an executable repository test SHALL fail if `README.md` omits those quick-reference entries
