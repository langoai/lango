
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

#### Scenario: Model role mapping
- **WHEN** a stored message has `Role: "model"` and empty `Author`
- **THEN** the EventsAdapter SHALL map the author to `rootAgentName` (or `"lango-agent"` if rootAgentName is empty)
- **AND** the author SHALL NOT be the literal string `"model"`

#### Scenario: Unknown role fallback
- **WHEN** a stored message has an unrecognized `Role` and empty `Author`
- **THEN** the EventsAdapter SHALL map the author to `rootAgentName` (or `"lango-agent"` if rootAgentName is empty)
- **AND** the author SHALL NOT produce "Event from an unknown agent" warnings

### Requirement: Agent hallucination retry in RunAndCollect
`RunAndCollect` SHALL detect "failed to find agent" errors, extract the hallucinated agent name, send a correction message with valid sub-agent names, and retry once. If the retry also fails, the original error SHALL be returned.

#### Scenario: Hallucinated agent name triggers retry
- **WHEN** a `RunAndCollect` call yields an error matching `"failed to find agent: <name>"`
- **AND** the agent has sub-agents registered
- **THEN** the system SHALL send a correction message: `[System: Agent "<name>" does not exist. Valid agents: <list>. Please retry using one of the valid agent names listed above.]`
- **AND** retry the run exactly once with the correction message

#### Scenario: Retry succeeds
- **WHEN** the correction message retry produces a successful response
- **THEN** `RunAndCollect` SHALL return the successful response with no error

#### Scenario: Retry also fails
- **WHEN** the correction message retry also produces an error
- **THEN** `RunAndCollect` SHALL return the retry error

#### Scenario: Non-hallucination error is not retried
- **WHEN** `RunAndCollect` encounters an error that does not match "failed to find agent"
- **THEN** the error SHALL be returned immediately without retry

#### Scenario: No sub-agents means no retry
- **WHEN** `RunAndCollect` encounters a "failed to find agent" error
- **AND** the agent has no sub-agents
- **THEN** the error SHALL be returned immediately without retry

### Requirement: Model adapter streams LLM responses
The `ModelAdapter.GenerateContent` method SHALL respect the `stream` parameter to control how `LLMResponse` events are yielded to the ADK runner.

When `stream` is `false`, the adapter SHALL consume all provider `StreamEvent`s internally, accumulate text and tool call parts, and yield exactly one `LLMResponse` with `Partial=false`, `TurnComplete=true`, and `Content.Parts` containing the full accumulated text and tool calls.

When `stream` is `true`, the adapter SHALL yield partial `LLMResponse` events for each text delta (`Partial=true`), and the final done event SHALL include the fully accumulated text in `Content.Parts` with `Partial=false` and `TurnComplete=true`.

#### Scenario: Non-streaming mode accumulates complete response
- **WHEN** `GenerateContent` is called with `stream=false` and the provider emits text deltas "Hello " and "world" followed by a done event
- **THEN** the adapter yields exactly one `LLMResponse` with `Partial=false`, `TurnComplete=true`, and `Content.Parts[0].Text` equal to "Hello world"

#### Scenario: Non-streaming mode accumulates tool calls
- **WHEN** `GenerateContent` is called with `stream=false` and the provider emits a tool call event followed by a done event
- **THEN** the adapter yields exactly one `LLMResponse` with `Partial=false`, `TurnComplete=true`, and `Content.Parts` containing the `FunctionCall`

#### Scenario: Streaming mode yields partial events and complete final
- **WHEN** `GenerateContent` is called with `stream=true` and the provider emits text deltas "Hello " and "world" followed by a done event
- **THEN** the adapter yields two partial `LLMResponse` events (one per delta) with `Partial=true`, followed by one final `LLMResponse` with `Partial=false`, `TurnComplete=true`, and `Content.Parts[0].Text` equal to "Hello world"

#### Scenario: Provider error propagates correctly
- **WHEN** the provider emits a `StreamEventError` event in either streaming or non-streaming mode
- **THEN** the adapter yields the error immediately and stops iteration

### Requirement: Agent text collection avoids duplication
The `runAndCollectOnce` method SHALL collect text from either partial events or the final non-partial event, but never both, to prevent duplicate text in the response.

#### Scenario: Streaming mode collects from partial events only
- **WHEN** `runAndCollectOnce` processes events that include partial text events followed by a non-partial final event containing the same accumulated text
- **THEN** the method returns text collected only from partial events, not from the final event

#### Scenario: Non-streaming mode collects from final event
- **WHEN** `runAndCollectOnce` processes events that contain no partial events and one non-partial final event with text
- **THEN** the method returns text from the non-partial final event

### Requirement: AppendEvent preserves FunctionCall metadata
The `SessionServiceAdapter.AppendEvent` method SHALL preserve the original `FunctionCall.ID` from the ADK event when storing to the internal session. When `FunctionCall.ID` is empty, it SHALL fall back to a synthetic ID of `"call_" + FunctionCall.Name`.

#### Scenario: FunctionCall with original ID
- **WHEN** an ADK event contains a `FunctionCall` with `ID: "adk-uuid-123"` and `Name: "exec"`
- **THEN** the stored `ToolCall.ID` SHALL be `"adk-uuid-123"` and `ToolCall.Name` SHALL be `"exec"`

#### Scenario: FunctionCall without ID
- **WHEN** an ADK event contains a `FunctionCall` with empty `ID` and `Name: "search"`
- **THEN** the stored `ToolCall.ID` SHALL be `"call_search"`

### Requirement: AppendEvent stores FunctionResponse metadata
The `SessionServiceAdapter.AppendEvent` method SHALL store `FunctionResponse` metadata (ID, Name, Response) in the message's `ToolCalls` array using the `Output` field for serialized response data. The response SHALL also be appended to `Content` for backward compatibility.

#### Scenario: FunctionResponse with full metadata
- **WHEN** an ADK event contains a `FunctionResponse` with `ID: "adk-uuid-123"`, `Name: "exec"`, and `Response: {"output": "file.txt"}`
- **THEN** the stored message SHALL have a `ToolCall` with `ID: "adk-uuid-123"`, `Name: "exec"`, and `Output` containing the serialized response JSON
- **AND** the message `Content` SHALL contain the serialized response JSON

### Requirement: EventsAdapter reconstructs FunctionCall with ID
The `EventsAdapter.All()` method SHALL reconstruct `genai.FunctionCall` parts with the stored `ID` field for assistant/model messages.

#### Scenario: Assistant message with FunctionCall ToolCalls
- **WHEN** an assistant message has `ToolCalls` with `ID: "adk-uuid-123"`, `Name: "exec"`, `Input: '{"cmd":"ls"}'`
- **THEN** the reconstructed event SHALL contain a `genai.FunctionCall` part with `ID: "adk-uuid-123"`, `Name: "exec"`, and parsed `Args`

### Requirement: EventsAdapter reconstructs FunctionResponse from new format
The `EventsAdapter.All()` method SHALL reconstruct `genai.FunctionResponse` parts from tool/function messages that have `ToolCalls` with `Output` data. The event role SHALL be set to `"function"`.

#### Scenario: Tool message with FunctionResponse metadata in ToolCalls
- **WHEN** a tool message has `ToolCalls` with `ID: "adk-uuid-123"`, `Name: "exec"`, `Output: '{"result":"ok"}'`
- **THEN** the reconstructed event SHALL contain a `genai.FunctionResponse` part with `ID: "adk-uuid-123"`, `Name: "exec"`, and parsed `Response`
- **AND** the event content role SHALL be `"function"`

### Requirement: EventsAdapter legacy FunctionResponse fallback
The `EventsAdapter.All()` method SHALL support legacy tool messages that lack `ToolCalls` metadata by inferring `FunctionResponse` from the preceding assistant message's `ToolCalls` using positional matching.

#### Scenario: Legacy tool message without ToolCalls
- **WHEN** a tool message has no `ToolCalls` but has `Content: '{"result":"file.txt"}'`
- **AND** the preceding assistant message has `ToolCalls` with `ID: "call_exec"`, `Name: "exec"`
- **THEN** the reconstructed event SHALL contain a `genai.FunctionResponse` with `ID: "call_exec"`, `Name: "exec"`, and the content parsed as `Response`

#### Scenario: Tool message with no context for FunctionResponse
- **WHEN** a tool message has no `ToolCalls` and no preceding assistant `ToolCalls`
- **THEN** the reconstructed event SHALL contain a text part with the message content

### Requirement: Token budget truncation preserves sequence safety
The `tokenBudgetTruncate` method SHALL ensure the truncated history does not start with a tool/function message or an orphaned assistant+FunctionCall without its matching response, but only when truncation actually removed messages.

#### Scenario: Truncation cuts before tool message
- **WHEN** token budget truncation removes earlier messages and the resulting slice starts with a tool/function message
- **THEN** the leading tool/function messages SHALL be skipped

#### Scenario: No truncation preserves trailing FunctionCall
- **WHEN** no truncation occurs (all messages fit within budget)
- **AND** the last message is an assistant with FunctionCall and no following tool response
- **THEN** the message SHALL NOT be removed (response is pending)

### Requirement: convertMessages forwards FunctionCall ID
The `convertMessages` function SHALL use the original `FunctionCall.ID` when converting `genai.Content` to `provider.Message`. When `FunctionCall.ID` is empty, it SHALL fall back to `"call_" + FunctionCall.Name`.

#### Scenario: FunctionCall with original ID in convertMessages
- **WHEN** a `genai.Content` contains a `FunctionCall` with `ID: "adk-uuid-123"`
- **THEN** the converted `provider.ToolCall.ID` SHALL be `"adk-uuid-123"`
