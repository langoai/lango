## ADDED Requirements
### Requirement: README automation completeness guard stays executable
Repository-level regressions that drop implemented `lango cron`, `lango workflow`, or `lango bg` command entries from the README quick reference SHALL be enforced by an executable test.

#### Scenario: Missing README automation entries are rejected
- **WHEN** the implemented automation CLI families remain shipped
- **THEN** an executable repository test SHALL fail if `README.md` omits those quick-reference entries
