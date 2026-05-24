## ADDED Requirements

### Requirement: Security quick-reference truthfulness guard stays executable
Repository-level regressions that blur the distinction between canonical `change-passphrase` and deprecated `migrate-passphrase` in public quick references SHALL be enforced by an executable test.

#### Scenario: Security quick references keep canonical/deprecated wording
- **WHEN** the repository still ships canonical `lango security change-passphrase` and deprecated `lango security migrate-passphrase`
- **THEN** an executable repository test SHALL fail if `README.md` or `docs/cli/index.md` stops distinguishing the non-reencrypting canonical path from the deprecated legacy migration path
