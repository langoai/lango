## MODIFIED Requirements

### Requirement: Store interface compatibility
The new ent-based store SHALL implement the existing `Store` interface. The ADK `SessionServiceAdapter` layer wrapping the store SHALL implement get-or-create semantics: when `Get()` returns a "session not found" error, the adapter SHALL call `Create()` to auto-create the session and return it.

#### Scenario: Interface compliance
- **WHEN** `EntStore` is used
- **THEN** it SHALL implement Create, Get, Update, Delete, AppendMessage, and Close methods

#### Scenario: ADK adapter auto-creation on miss
- **WHEN** `SessionServiceAdapter.Get()` receives a "session not found" error from the store
- **THEN** it SHALL call `SessionServiceAdapter.Create()` with the requested session ID and return the newly created session
