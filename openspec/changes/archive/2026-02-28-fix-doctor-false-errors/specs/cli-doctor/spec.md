## MODIFIED Requirements

### Requirement: Graph store health check
The doctor command SHALL include a GraphStoreCheck that validates graph store configuration. The check SHALL skip if graph.enabled is false. When enabled, it SHALL validate that backend is "bolt" and maxTraversalDepth and maxExpansionResults are positive. When databasePath is empty, the check SHALL return StatusWarn with a message indicating the path will default to graph.db next to the session database, instead of StatusFail.

#### Scenario: Graph disabled
- **WHEN** doctor runs with graph.enabled=false
- **THEN** GraphStoreCheck returns StatusSkip

#### Scenario: Graph databasePath empty
- **WHEN** doctor runs with graph.enabled=true and databasePath empty
- **THEN** GraphStoreCheck returns StatusWarn with message indicating the fallback path will be used

#### Scenario: Graph misconfigured backend
- **WHEN** doctor runs with graph.enabled=true and backend is not "bolt"
- **THEN** GraphStoreCheck returns StatusFail with message about unsupported backend

### Requirement: Session Database Check
The system SHALL verify that the session database is accessible. The fallback database path when no config is loaded SHALL be `~/.lango/lango.db`, matching the DefaultConfig convention.

#### Scenario: Database file exists and is writable
- **WHEN** session.databasePath points to an accessible SQLite file
- **THEN** check passes with database path displayed

#### Scenario: Database path not writable
- **WHEN** database path directory is not writable
- **THEN** check fails with permission error

#### Scenario: No config loaded fallback path
- **WHEN** no configuration is loaded (cfg is nil or databasePath is empty)
- **THEN** the check SHALL use `~/.lango/lango.db` as the fallback path
