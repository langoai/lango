## ADDED Requirements

### Requirement: Deprecated SQLCipher open-arg behavior stays executable
Repository-level regressions that make legacy encryption arguments behave as active SQLCipher controls again SHALL be prevented by executable package tests around the DB-open paths.

#### Scenario: Plaintext DB-open paths ignore deprecated encryption args
- **WHEN** the repository still ships managed and read-only DB-open paths without SQLCipher runtime support
- **THEN** executable package tests SHALL fail if plaintext managed or read-only opens start failing solely because deprecated encryption arguments are provided
