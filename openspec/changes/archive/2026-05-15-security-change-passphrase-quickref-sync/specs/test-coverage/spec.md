## ADDED Requirements

### Requirement: Security quick-reference completeness guard stays executable
Repository-level regressions that drop implemented `lango security` command entries from the public quick references SHALL be enforced by an executable test.

#### Scenario: Implemented security quick-reference entries remain listed
- **WHEN** the repository still ships the implemented `lango security` command family, including canonical `change-passphrase`
- **THEN** an executable repository test SHALL fail if `README.md` or `docs/cli/index.md` omits those quick-reference entries
