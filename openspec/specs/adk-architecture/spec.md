
## ADDED Requirements

### Requirement: ADK Agent Abstraction
The system SHALL wrap the Google ADK Agent to integrate with the application.

#### Scenario: Agent Initialization
- **WHEN** the server starts
- **THEN** it SHALL initialize an ADK Agent instance
- **AND** configure it with the selected model and tools from the configuration

### Requirement: Ent State Adapter
The system SHALL adapt the Ent-based session store to the ADK State interface.

#### Scenario: Load Session
- **WHEN** ADK requests state for a session ID
- **THEN** the adapter SHALL retrieve the session and messages from Ent
- **AND** map them to ADK's expected message format

#### Scenario: Save Session
- **WHEN** ADK persists state updates
- **THEN** the adapter SHALL save new messages and state to Ent
- **AND** the in-memory session history SHALL be updated to reflect the persisted message

### Requirement: AppendEvent In-Memory History Sync
The system SHALL update the in-memory session history when appending events, in addition to persisting to the database store.

#### Scenario: User message visible after AppendEvent
- **WHEN** `SessionServiceAdapter.AppendEvent` is called with a user message event
- **THEN** the message SHALL be persisted to the database store
- **AND** the message SHALL be appended to `SessionAdapter.sess.History` in memory
- **AND** subsequent calls to `session.Events().All()` on the same session object SHALL include the new message

#### Scenario: Multiple events accumulate in memory
- **WHEN** multiple events are appended to the same session in sequence
- **THEN** all messages SHALL be visible in the in-memory history in order of insertion

#### Scenario: State-delta-only events skip history
- **WHEN** an event contains only `Actions.StateDelta` with no `LLMResponse.Content`
- **THEN** the event SHALL NOT be appended to the in-memory history
- **AND** the event SHALL NOT be persisted to the database store

### Requirement: SystemInstruction Forwarding
The system SHALL forward ADK SystemInstruction to the LLM provider as a system-role message.

#### Scenario: SystemInstruction present in request
- **WHEN** `ModelAdapter.GenerateContent` receives a request with `req.Config.SystemInstruction` containing text parts
- **THEN** the text parts SHALL be concatenated and prepended as a `provider.Message{Role: "system"}` before all other messages

#### Scenario: SystemInstruction absent
- **WHEN** `ModelAdapter.GenerateContent` receives a request without `SystemInstruction` (nil Config or nil SystemInstruction)
- **THEN** no system message SHALL be prepended
- **AND** only the original `req.Contents` SHALL be forwarded to the provider

#### Scenario: Multi-part SystemInstruction
- **WHEN** `SystemInstruction` contains multiple text parts
- **THEN** all parts SHALL be joined with newline separators into a single system message

### Requirement: Tool Adapter
The system SHALL adapt existing internal tools to the ADK Tool interface.

#### Scenario: Execute Legacy Tool
- **WHEN** ADK invokes a tool
- **THEN** the adapter SHALL translate the inputs and call the internal tool implementation
### Requirement: History Management
The system SHALL manage session history using token-budget-based dynamic truncation to prevent context overflow and optimize token usage.

#### Scenario: History Truncation with explicit budget
- **WHEN** loading session history for the agent with an explicit token budget
- **THEN** the token budget SHALL be applied
- **AND** messages SHALL be included from most recent to oldest until the budget is exhausted
- **AND** any remaining older messages SHALL be excluded from the LLM context

#### Scenario: Default token budget
- **WHEN** no explicit token budget is provided (budget = 0)
- **THEN** the system SHALL use a default budget of 32000 tokens

#### Scenario: Event Author Mapping
- **WHEN** adapting historical messages to ADK events
- **THEN** the `Author` field SHALL be populated based on the message role
- **AND** `user` role maps to `user` author
- **AND** `assistant` role maps to the agent name
