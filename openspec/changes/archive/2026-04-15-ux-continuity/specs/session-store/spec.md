## ADDED Requirements

### Requirement: Session end API
The `session.Store` interface SHALL expose an `End(key string) error` method that marks a session as ended. Ending a session SHALL set metadata key `lango.session_end_pending=true` and trigger the configured session-end processor (see `session-recall` capability). Calling `End` on an already-ended session SHALL be a no-op.

#### Scenario: End marks metadata
- **WHEN** `store.End("sess-1")` is called on an active session
- **THEN** the session's metadata SHALL contain `lango.session_end_pending=true`

#### Scenario: End is idempotent
- **WHEN** `store.End("sess-1")` is called twice
- **THEN** the second call SHALL return `nil` without error
- **AND** metadata SHALL remain stable

#### Scenario: End on unknown session returns error
- **WHEN** `store.End("missing")` is called where the session does not exist
- **THEN** a session-not-found error SHALL be returned

### Requirement: Session-end pending flag
The system SHALL use the metadata key `lango.session_end_pending` (boolean, serialized as string `"true"`/`"false"`) to mark sessions that have a pending recall-indexing job. The store SHALL expose helpers `MarkEndPending(key)`, `ClearEndPending(key)`, and `ListEndPending()` returning keys with the flag set.

#### Scenario: ListEndPending returns pending sessions
- **WHEN** two sessions have `lango.session_end_pending=true` and one has it cleared
- **THEN** `ListEndPending()` SHALL return the two pending keys and not include the cleared one

#### Scenario: ClearEndPending flips the flag
- **WHEN** `ClearEndPending("sess-1")` is called on a pending session
- **THEN** subsequent `ListEndPending()` calls SHALL NOT include `sess-1`

### Requirement: Session-end processor hook
The system SHALL allow registering a `SessionEndProcessor` function (accepting a session key and returning an error) via `session.Store.SetSessionEndProcessor`. The store SHALL invoke the processor when `End(key)` is called (hard-end path) bounded by a caller-supplied timeout, and sweeps MAY invoke the processor for pending sessions asynchronously (soft-end recovery path).

#### Scenario: Hard end invokes processor synchronously with timeout
- **WHEN** `End("sess-1")` is called with a 3s bound and a processor is registered
- **THEN** the processor SHALL be invoked with key `"sess-1"`
- **AND** the call SHALL return within the 3s bound even if the processor is still running (timeout case leaves `lango.session_end_pending=true`)

#### Scenario: Sweep invokes processor for pending sessions
- **WHEN** a sweep runs and finds `sess-1` with `lango.session_end_pending=true`
- **THEN** the processor SHALL be invoked asynchronously
- **AND** on success `ClearEndPending("sess-1")` SHALL be called

#### Scenario: No processor registered is a no-op
- **WHEN** `End("sess-1")` is called and no processor is registered
- **THEN** metadata SHALL still be set to `lango.session_end_pending=true`
- **AND** no error SHALL be returned
