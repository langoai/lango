## ADDED Requirements
### Requirement: README provenance completeness guard stays executable
Repository-level regressions that drop implemented `lango provenance` command entries from the README quick reference SHALL be enforced by an executable test.

#### Scenario: Missing README provenance entries are rejected
- **WHEN** the implemented `lango provenance` CLI family remains shipped
- **THEN** an executable repository test SHALL fail if `README.md` omits those quick-reference entries
