## MODIFIED Requirements

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
