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
