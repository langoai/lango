### Requirement: Auto-create session on first access
The `SessionServiceAdapter.Get()` SHALL automatically create a new session when the requested session ID does not exist in the store, instead of returning an error.

#### Scenario: First message from a new Telegram user
- **WHEN** `SessionServiceAdapter.Get()` is called with a session ID that does not exist (e.g., `telegram:123:456`)
- **THEN** the system SHALL create a new session with that key and return it successfully

#### Scenario: First message from a new Discord user
- **WHEN** `SessionServiceAdapter.Get()` is called with a session ID that does not exist (e.g., `discord:chan1:user1`)
- **THEN** the system SHALL create a new session with that key and return it successfully

#### Scenario: First message from a new Slack user
- **WHEN** `SessionServiceAdapter.Get()` is called with a session ID that does not exist (e.g., `slack:chan1:user1`)
- **THEN** the system SHALL create a new session with that key and return it successfully

#### Scenario: Gateway default session
- **WHEN** `SessionServiceAdapter.Get()` is called with session ID `default` that does not exist
- **THEN** the system SHALL create a new session with key `default` and return it successfully

### Requirement: Existing sessions retrieved normally
The `SessionServiceAdapter.Get()` SHALL return existing sessions without creating duplicates. When provenance is enabled, Get() SHALL ensure the session is registered in the provenance tree (missing-only backfill).

#### Scenario: Subsequent messages from an existing user
- **WHEN** `SessionServiceAdapter.Get()` is called with a session ID that already exists
- **THEN** the system SHALL return the existing session with its conversation history intact

#### Scenario: Existing session backfilled in provenance tree
- **WHEN** `Get()` returns an existing session
- **AND** the session is not yet registered in the provenance tree
- **THEN** the rootSessionObserver SHALL register the session (missing-only backfill)

#### Scenario: Already registered session not re-registered
- **WHEN** `Get()` returns an existing session
- **AND** the session is already registered in the provenance tree
- **THEN** the rootSessionObserver SHALL skip registration (idempotent)

### Requirement: Non-recoverable store errors propagated
The `SessionServiceAdapter.Get()` SHALL propagate store errors that are not "session not found" or "session expired" (e.g., database connection failures).

#### Scenario: Database error during get
- **WHEN** the store returns an error other than "session not found" or "session expired"
- **THEN** the system SHALL propagate that error to the caller without attempting auto-creation or renewal

### Requirement: Concurrent auto-create safety
The `SessionServiceAdapter.Get()` SHALL handle concurrent auto-creation attempts for the same session key without returning errors. When multiple goroutines simultaneously detect a missing session and attempt creation, at most one SHALL succeed in creating it, and the others SHALL retrieve the already-created session.

#### Scenario: Multiple Telegram messages arrive simultaneously for a new user
- **WHEN** multiple goroutines call `SessionServiceAdapter.Get()` concurrently with the same non-existent session ID
- **THEN** all goroutines SHALL return the session successfully, and exactly one session SHALL exist in the store

#### Scenario: Create fails with UNIQUE constraint
- **WHEN** `SessionServiceAdapter.getOrCreate()` attempts to create a session and the store returns a UNIQUE constraint error
- **THEN** the method SHALL retry `store.Get()` to fetch the session created by another goroutine and return it successfully

#### Scenario: Create fails with non-constraint error
- **WHEN** `SessionServiceAdapter.getOrCreate()` attempts to create a session and the store returns an error that is not a UNIQUE constraint violation
- **THEN** the method SHALL propagate the error to the caller

### Requirement: Auto-renew expired sessions
The `SessionServiceAdapter.Get()` SHALL automatically delete an expired session and create a fresh replacement when the store returns `ErrSessionExpired`, so the user's current message is processed normally.

#### Scenario: Expired Telegram session auto-renews
- **WHEN** `SessionServiceAdapter.Get()` receives `ErrSessionExpired` for session `telegram:123:456`
- **THEN** the system SHALL delete the expired session, create a new session with the same key, and return it successfully

#### Scenario: Expired session delete failure propagates error
- **WHEN** `SessionServiceAdapter.Get()` receives `ErrSessionExpired` and the subsequent `Delete()` call fails
- **THEN** the system SHALL return the delete error wrapped with context, without attempting to create a new session

#### Scenario: Concurrent expiry recovery is safe
- **WHEN** multiple goroutines detect the same expired session simultaneously
- **THEN** the `getOrCreate()` retry logic SHALL ensure all goroutines return a valid session without errors
