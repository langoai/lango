## ADDED Requirements

### Requirement: SQLite driver and db-open regression coverage stays executable
Repository-level regressions that break low-level SQLite driver helpers or managed/read-only database opening flows SHALL be enforced by executable tests close to those packages.

#### Scenario: SQLite helper and DB-open paths remain covered
- **WHEN** the repository still ships `internal/sqlitedriver` and `internal/dbopen`
- **THEN** executable package tests SHALL cover DSN construction, file-header validation including legacy unreadable-header rejection, and connection configuration in `internal/sqlitedriver`
- **AND** executable package tests SHALL cover managed DB creation, read-only reopen of an existing managed DB, missing-file read-only failure, and legacy-header fail-fast in `internal/dbopen`
