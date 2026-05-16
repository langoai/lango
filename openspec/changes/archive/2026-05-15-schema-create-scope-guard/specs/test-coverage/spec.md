## ADDED Requirements

### Requirement: Production Schema.Create scope and serialization stay executable
Repository-level regressions that reintroduce unsynchronized production Ent schema-migration call sites SHALL be enforced by an executable quality guard.

#### Scenario: Production Schema.Create remains scoped to serialized constructors
- **WHEN** the repository still ships production `Schema.Create` call sites
- **THEN** an executable repository test SHALL fail if non-test `Schema.Create` usage appears outside the approved runtime-owned constructors
- **AND** it SHALL fail if the approved constructors stop holding `schemaCreateMu` around those `Schema.Create` calls
