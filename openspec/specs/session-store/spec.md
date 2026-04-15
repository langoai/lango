# Session Store Specification

## Purpose
The session store manages conversation sessions, message history, and metadata persistence using a robust, type-safe database backend.
## Requirements
### Requirement: Session creation
The system SHALL create new sessions with unique identifiers and store them in SQLite.

#### Scenario: Create new session
- **WHEN** a new conversation begins
- **THEN** a session record SHALL be created with a unique key

#### Scenario: Session with agent assignment
- **WHEN** creating a session for a specific agent
- **THEN** the agent ID SHALL be associated with the session

### Requirement: Message structure
The `session.Message` struct SHALL use `types.MessageRole` for its `Role` field instead of plain `string`. All internal code that reads or writes `Message.Role` SHALL use typed enum constants (`types.RoleUser`, `types.RoleAssistant`, `types.RoleTool`, `types.RoleFunction`, `types.RoleModel`). The `string()` cast SHALL only occur at system boundaries: Ent DB writes (`SetRole(string(msg.Role))`), Ent DB reads (`types.MessageRole(m.Role)`), and external API mapping (genai `Content.Role`).

#### Scenario: Message role uses typed enum
- **WHEN** a `session.Message` is created anywhere in internal code
- **THEN** the `Role` field SHALL be assigned a `types.MessageRole` constant, not a raw string literal

#### Scenario: DB boundary cast on write
- **WHEN** a message is persisted to the Ent store via `SetRole()`
- **THEN** the role SHALL be cast to `string` at the call site: `SetRole(string(msg.Role))`

#### Scenario: DB boundary cast on read
- **WHEN** a message is loaded from the Ent store
- **THEN** the role SHALL be cast from `string` to `types.MessageRole`: `Role: types.MessageRole(m.Role)`

#### Scenario: JSON serialization backward compatibility
- **WHEN** a `session.Message` with `Role: types.RoleUser` is serialized to JSON
- **THEN** the JSON output SHALL contain `"role":"user"` (unchanged from previous format)

### Requirement: Message history storage
The system SHALL store conversation message history in the session.

#### Scenario: Store user message
- **WHEN** a user message is processed
- **THEN** the message SHALL be appended to the session history

#### Scenario: Store assistant response
- **WHEN** the assistant generates a response
- **THEN** the response SHALL be appended to the session history

### Requirement: Session retrieval
The system SHALL retrieve session data including message history.

#### Scenario: Load session by key
- **WHEN** a session key is provided
- **THEN** the full session data SHALL be loaded

#### Scenario: Session not found
- **WHEN** an invalid session key is provided
- **THEN** a session-not-found error SHALL be returned

### Requirement: Session metadata
The system SHALL store and retrieve session metadata (model, settings).

#### Scenario: Store session settings
- **WHEN** session settings are updated (model, thinking level)
- **THEN** the settings SHALL be persisted

#### Scenario: Retrieve session settings
- **WHEN** a session is loaded
- **THEN** the current settings SHALL be included

### Requirement: Session cleanup
The system SHALL support session deletion and expiration.

#### Scenario: Delete session
- **WHEN** session deletion is requested
- **THEN** all session data SHALL be removed from storage

#### Scenario: Session expiration
- **WHEN** a session exceeds its TTL
- **THEN** the session MAY be marked for cleanup

### Requirement: Session storage implementation
The session store implementation SHALL use entgo.io instead of raw SQL queries.

#### Scenario: Create session implementation
- **WHEN** `Store.Create(session)` is called
- **THEN** the session SHALL be persisted using ent client

#### Scenario: Get session implementation
- **WHEN** `Store.Get(key)` is called
- **THEN** the session SHALL be retrieved using ent query with Message eager loading

#### Scenario: AppendMessage implementation
- **WHEN** `Store.AppendMessage(key, msg)` is called
- **THEN** a new Message entity SHALL be created linked to the Session

### Requirement: Message Author Field
The `session.Message` struct SHALL include an `Author string` field (JSON tag `"author,omitempty"`) to store the ADK agent name that produced the message.

#### Scenario: Author preserved through AppendEvent
- **WHEN** an ADK event with `Author: "lango-orchestrator"` is appended
- **THEN** the stored message SHALL have `Author: "lango-orchestrator"`

#### Scenario: Author loaded from storage
- **WHEN** a session is loaded from the ent store
- **THEN** each message's Author field SHALL be populated from the stored `author` column

### Requirement: ToolCall stores FunctionResponse output
The `ToolCall` struct's `Output` field SHALL be used to store serialized `FunctionResponse.Response` data for tool/function role messages. This enables round-trip preservation of FunctionResponse metadata through the save/restore cycle without database schema changes.

#### Scenario: Tool message ToolCall with Output
- **WHEN** a tool message is stored with `ToolCalls` containing `Output` data
- **AND** the message is later loaded from the database
- **THEN** the `ToolCall.Output` field SHALL contain the original serialized response JSON
- **AND** `ToolCall.ID` and `ToolCall.Name` SHALL match the original FunctionResponse metadata

### Requirement: ToolCall persists thinking metadata
The `session.ToolCall` and `entschema.ToolCall` structs SHALL include `Thought bool` and `ThoughtSignature []byte` fields with `omitempty` JSON tags. These fields SHALL survive the full persistence round-trip: session → database → session reload.

#### Scenario: Persist and reload ThoughtSignature
- **WHEN** a session message with FunctionCall ToolCalls containing `ThoughtSignature` is persisted via `AppendMessage`
- **THEN** retrieving the session via `Get` SHALL return ToolCalls with the original `Thought` and `ThoughtSignature` values intact

#### Scenario: Legacy session without thinking fields
- **WHEN** an existing session record has ToolCalls without `thought` or `thoughtSignature` JSON keys
- **THEN** deserialization SHALL produce `Thought=false` and `ThoughtSignature=nil` (zero values)

### Requirement: Empty response fallback for channel path
The `runAgent` function in `channels.go` SHALL return a user-visible fallback message when the agent succeeds (no error) but produces an empty response string. The fallback message SHALL be a package-level constant.

#### Scenario: Agent returns empty response via channel
- **WHEN** the agent `RunAndCollect` returns an empty string with no error
- **THEN** `runAgent` SHALL substitute the `emptyResponseFallback` constant as the response
- **AND** SHALL log a warning with session key and elapsed time

#### Scenario: Agent returns non-empty response via channel
- **WHEN** the agent `RunAndCollect` returns a non-empty string with no error
- **THEN** `runAgent` SHALL return the response unchanged

### Requirement: Empty response fallback for gateway path
The `handleChatMessage` function in `gateway/server.go` SHALL return a user-visible fallback message when the agent succeeds but produces an empty response string via the WebSocket streaming path.

#### Scenario: Agent returns empty response via gateway
- **WHEN** the agent `RunStreaming` returns an empty string with no error
- **THEN** `handleChatMessage` SHALL substitute the `emptyResponseFallback` constant as the response
- **AND** SHALL log a warning with session key

#### Scenario: Agent returns error via gateway
- **WHEN** the agent `RunStreaming` returns an error
- **THEN** `handleChatMessage` SHALL NOT apply the fallback and SHALL propagate the error normally

### Requirement: Agent empty response diagnostic logging
The `Agent.RunAndCollect` function SHALL log a warning when the agent run succeeds but produces an empty response string, providing session ID and elapsed time for diagnostics.

#### Scenario: Empty response logged at agent level
- **WHEN** `runAndCollectOnce` returns an empty string with no error
- **THEN** `RunAndCollect` SHALL log a warn-level message with session and elapsed fields

### Requirement: Agent text collection without thought filter
The `runAndCollectOnce` and `RunStreaming` functions SHALL collect text from session event parts using `part.Text != ""` without filtering on `part.Thought`. Thought filtering is the responsibility of the provider layer, not the agent layer.

#### Scenario: Text parts collected without thought check
- **WHEN** a session event contains text parts
- **THEN** the agent SHALL collect all non-empty text parts regardless of the `Thought` field value

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

