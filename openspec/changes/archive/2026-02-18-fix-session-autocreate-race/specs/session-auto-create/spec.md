## ADDED Requirements

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
