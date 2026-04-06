## Purpose

The Persistent Approval Grant capability provides an in-memory, per-session, per-tool "always allow" grant store. When a user clicks "Always Allow" for a tool, subsequent invocations of the same tool within the same session are auto-approved without re-prompting. Grants are cleared on application restart.

## Requirements

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

### Requirement: Grant TTL expiration
GrantStore SHALL support an optional time-to-live (TTL) for grants. When TTL is set to a positive duration, `IsGranted()` MUST check whether the grant has expired (current time minus `grantedAt` exceeds TTL). A TTL of zero MUST preserve backward-compatible behavior (no expiration).

#### Scenario: Grant within TTL
- **WHEN** a grant was created 5 minutes ago and TTL is 10 minutes
- **THEN** `IsGranted()` SHALL return true

#### Scenario: Grant expired past TTL
- **WHEN** a grant was created 11 minutes ago and TTL is 10 minutes
- **THEN** `IsGranted()` SHALL return false

#### Scenario: TTL zero means no expiry
- **WHEN** TTL is zero (default) and a grant was created 100 hours ago
- **THEN** `IsGranted()` SHALL return true

### Requirement: Clean expired grants
GrantStore SHALL provide a `CleanExpired()` method that removes all grants whose `grantedAt` timestamp exceeds the configured TTL. The method SHALL return the count of removed entries. When TTL is zero, `CleanExpired()` SHALL be a no-op returning zero.

#### Scenario: Clean expired entries
- **WHEN** `CleanExpired()` is called with TTL of 5 minutes and 2 of 3 grants are older than 5 minutes
- **THEN** the method SHALL remove the 2 expired grants and return 2

#### Scenario: Clean with zero TTL
- **WHEN** `CleanExpired()` is called with TTL of zero
- **THEN** the method SHALL remove nothing and return 0

### Requirement: P2P grant TTL default
When P2P is enabled, the application SHALL set the GrantStore TTL to 1 hour. This limits the window of implicit trust from P2P approval grants.

#### Scenario: P2P enabled sets 1-hour TTL
- **WHEN** the application initializes with `cfg.P2P.Enabled = true`
- **THEN** `grantStore.SetTTL(time.Hour)` SHALL be called

### Requirement: Double-approval prevention via grant recording
When the P2P approval function approves a tool invocation, the system SHALL record a grant for `"p2p:"+peerDID` and the tool name. This prevents the tool's internal `wrapWithApproval` from prompting a second time.

#### Scenario: Approved P2P tool records grant
- **WHEN** the P2P approval function approves tool "echo" for peer "did:key:abc"
- **THEN** a grant SHALL be recorded with session key `"p2p:did:key:abc"` and tool name `"echo"`

### Requirement: Grant store enumeration
The GrantStore SHALL provide a List() method returning all active (non-expired) grants as []GrantInfo sorted by session key then tool name. Expired grants SHALL be excluded.

#### Scenario: List active grants
- **WHEN** 3 grants exist and none are expired
- **THEN** List() returns 3 GrantInfo sorted by session then tool

#### Scenario: Expired grants excluded
- **WHEN** a grant has expired based on the TTL
- **THEN** List() excludes it from the result

#### Scenario: Revoked grants excluded
- **WHEN** a grant has been revoked
- **THEN** List() excludes it from the result
