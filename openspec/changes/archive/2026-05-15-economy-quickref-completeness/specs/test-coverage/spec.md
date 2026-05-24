## ADDED Requirements
### Requirement: Economy quick-reference completeness guard stays executable
Repository-level regressions that drop implemented `economy escrow list/show/sentinel status` commands from the public quick references SHALL be enforced by an executable test.

#### Scenario: Missing economy escrow quick-reference entries are rejected
- **WHEN** those economy escrow commands remain shipped
- **THEN** an executable repository test SHALL fail if `README.md` or `docs/cli/index.md` omits them
