## ADDED Requirements

### Requirement: Broker transport regression tests remain release gates
The broker storage boundary MUST continue to verify transport correctness for concurrent writes, large JSON responses, and graceful shutdown as part of standard test execution.

#### Scenario: Broker transport regressions are exercised in standard test runs
- **WHEN** `go test ./...` is executed
- **THEN** broker transport regression tests cover concurrent request serialization, responses larger than 64 KiB, and shutdown-before-close behavior

### Requirement: Payment mutator setup no longer falls back to direct Ent extraction
Production payment setup MUST use storage-facing transaction persistence and limiter capabilities without falling back to `session.EntStore.Client()`.

#### Scenario: Payment setup resolved through storage capability only
- **WHEN** app or CLI payment setup initializes payment transaction persistence
- **THEN** it resolves the collaborator through storage-facing capabilities
- **AND** it does not rebuild the dependency from a session-store Ent client
