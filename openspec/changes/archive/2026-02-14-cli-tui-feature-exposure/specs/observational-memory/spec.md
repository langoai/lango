## ADDED Requirements

### Requirement: Delete reflections by session
The memory Store SHALL provide a `DeleteReflectionsBySession(ctx, sessionKey)` method that deletes all reflections for a given session key. The method SHALL follow the same pattern as `DeleteObservationsBySession`. The method SHALL return nil when the session has no reflections (no-op delete).

#### Scenario: Delete all reflections for a session
- **WHEN** `DeleteReflectionsBySession` is called with a session key that has reflections
- **THEN** all reflections for that session are deleted and reflections in other sessions are unaffected

#### Scenario: Delete from empty session
- **WHEN** `DeleteReflectionsBySession` is called with a session key that has no reflections
- **THEN** the method returns nil without error
