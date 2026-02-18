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
The `SessionServiceAdapter.Get()` SHALL return existing sessions without creating duplicates.

#### Scenario: Subsequent messages from an existing user
- **WHEN** `SessionServiceAdapter.Get()` is called with a session ID that already exists
- **THEN** the system SHALL return the existing session with its conversation history intact

### Requirement: Non-recoverable store errors propagated
The `SessionServiceAdapter.Get()` SHALL propagate store errors that are not "session not found" (e.g., database connection failures).

#### Scenario: Database error during get
- **WHEN** the store returns an error other than "session not found"
- **THEN** the system SHALL propagate that error to the caller without attempting auto-creation

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
