## ADDED Requirements

### Requirement: GrantStore provides thread-safe in-memory grant tracking
The system SHALL provide a `GrantStore` type that stores per-session, per-tool approval grants in memory using a `sync.RWMutex`-protected map.

#### Scenario: Grant and check a tool
- **WHEN** `Grant("session-1", "exec")` is called
- **THEN** `IsGranted("session-1", "exec")` SHALL return `true`

#### Scenario: Grants are isolated by session and tool
- **WHEN** `Grant("session-1", "exec")` is called
- **THEN** `IsGranted("session-1", "fs_delete")` SHALL return `false`
- **AND** `IsGranted("session-2", "exec")` SHALL return `false`

### Requirement: GrantStore supports single-tool revocation
The system SHALL provide a `Revoke(sessionKey, toolName)` method that removes a single grant without affecting other grants.

#### Scenario: Revoke one tool grant
- **WHEN** `Grant("s1", "exec")` and `Grant("s1", "fs_write")` are called, then `Revoke("s1", "exec")`
- **THEN** `IsGranted("s1", "exec")` SHALL return `false`
- **AND** `IsGranted("s1", "fs_write")` SHALL return `true`

### Requirement: GrantStore supports session-wide revocation
The system SHALL provide a `RevokeSession(sessionKey)` method that removes all grants for a given session.

#### Scenario: Revoke all grants for a session
- **WHEN** `Grant("s1", "exec")` and `Grant("s1", "fs_write")` and `Grant("s2", "exec")` are called, then `RevokeSession("s1")`
- **THEN** `IsGranted("s1", "exec")` SHALL return `false`
- **AND** `IsGranted("s1", "fs_write")` SHALL return `false`
- **AND** `IsGranted("s2", "exec")` SHALL return `true`

### Requirement: GrantStore is safe for concurrent access
The system SHALL allow concurrent `Grant`, `IsGranted`, `Revoke`, and `RevokeSession` calls without data races.

#### Scenario: Concurrent grant and check
- **WHEN** 100 goroutines concurrently call `Grant` and `IsGranted`
- **THEN** no data race SHALL occur and the final state SHALL be consistent
