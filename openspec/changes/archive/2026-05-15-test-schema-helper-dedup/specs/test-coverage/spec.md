## ADDED Requirements

### Requirement: Shared test schema helper stays cycle-safe and reusable
Repository-level regressions that reintroduce duplicated or import-cycle-prone Ent schema bootstrap logic in tests SHALL be prevented by a shared test-only helper boundary.

#### Scenario: Serialized schema setup remains reusable from tests
- **WHEN** tests in multiple packages need Ent schema bootstrap
- **THEN** they SHALL be able to use a minimal helper under `internal/testutil/schemautil` without importing broader testutil surfaces that create package cycles
- **AND** the helper SHALL continue to serialize Atlas-backed schema creation for those tests
