## ADDED Requirements

### Requirement: README security inventory truthfulness guard stays executable
Repository-level regressions that let the README internal CLI inventory collapse the current security command surface into stale shorthand SHALL be enforced by an executable test.

#### Scenario: README security inventory remains truthful
- **WHEN** the repository still ships canonical `lango security change-passphrase`, deprecated `migrate-passphrase`, `recovery setup/restore`, and `kms wrap/detach`
- **THEN** an executable repository test SHALL fail if `README.md` stops describing those current security command surfaces in its internal tree inventory
