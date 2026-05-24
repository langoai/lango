## ADDED Requirements

### Requirement: Direct test Schema.Create usage stays blocked
Repository-level regressions that let test code bypass the shared schema helper and call Ent schema creation directly SHALL be prevented by an executable quality guard.

#### Scenario: Tests keep using the shared schema helper
- **WHEN** test code under `internal/` still needs Ent schema bootstrap
- **THEN** an executable repository test SHALL fail if a test file reintroduces direct `Schema.Create` usage outside the approved `internal/testutil/schemautil` helper and the guard test that intentionally scans for that token
