## MODIFIED Requirements

### Requirement: Get auto-create for new sessions
The `SessionServiceAdapter.Get()` method SHALL auto-create missing sessions and ensure the provenance tree is backfilled for existing sessions that are not yet registered.

#### Scenario: Existing session backfilled in provenance tree
- **WHEN** `Get()` returns an existing session
- **AND** the session is not yet registered in the provenance tree
- **THEN** the rootSessionObserver SHALL register the session (missing-only backfill)

#### Scenario: Already registered session not re-registered
- **WHEN** `Get()` returns an existing session
- **AND** the session is already registered in the provenance tree
- **THEN** the rootSessionObserver SHALL skip registration (idempotent)
