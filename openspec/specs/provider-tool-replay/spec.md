## Purpose

Provider-specific tool call replay stabilization. Covers metadata preservation, ID propagation, backward compatibility for legacy sessions, orphaned tool call repair (OpenAI-specific), and corrupted thinking entry defense (Gemini-specific).

## Requirements

### Requirement: tool_call_name metadata preservation
The system SHALL store `tool_call_name` in FunctionResponse metadata alongside `tool_call_id` during message conversion from genai.Content to provider.Message.

#### Scenario: FunctionResponse round-trip preserves name
- **WHEN** a genai.Content with FunctionResponse (Name="exec", ID="call_1") is converted via convertMessages()
- **THEN** the resulting provider.Message metadata SHALL contain both `tool_call_id`="call_1" and `tool_call_name`="exec"

### Requirement: Gemini FunctionCall.ID propagation
The system SHALL set `FunctionCall.ID` from provider.ToolCall.ID when building Gemini genai.Content for assistant messages.

#### Scenario: FunctionCall includes ID in Gemini content
- **WHEN** an assistant message with ToolCall (ID="call_abc", Name="exec") is converted to Gemini content
- **THEN** the genai.Part.FunctionCall SHALL have ID="call_abc" and Name="exec"

### Requirement: Gemini FunctionResponse.ID propagation
The system SHALL set `FunctionResponse.ID` from metadata `tool_call_id` when building Gemini genai.Content for tool responses.

#### Scenario: FunctionResponse includes ID in Gemini content
- **WHEN** a tool message with metadata tool_call_id="call_abc" and tool_call_name="exec" is converted to Gemini content
- **THEN** the genai.Part.FunctionResponse SHALL have ID="call_abc" and Name="exec"

### Requirement: Gemini streaming FunctionCall.ID preference
The system SHALL use `FunctionCall.ID` from Gemini streaming responses when available, falling back to `FunctionCall.Name` only when ID is empty.

#### Scenario: Streaming response uses FunctionCall.ID when present
- **WHEN** a Gemini streaming response contains FunctionCall with ID="fc_123" and Name="exec"
- **THEN** the emitted provider.ToolCall SHALL have ID="fc_123"

#### Scenario: Streaming response falls back to Name when ID empty
- **WHEN** a Gemini streaming response contains FunctionCall with ID="" and Name="exec"
- **THEN** the emitted provider.ToolCall SHALL have ID="exec"

### Requirement: Gemini tool_call_name backward compatibility
The system SHALL infer `tool_call_name` from the nearest preceding assistant message's ToolCalls when the metadata field is missing (legacy sessions).

#### Scenario: Legacy session without tool_call_name infers from history
- **WHEN** a tool message has metadata tool_call_id="call_1" but no tool_call_name, AND the preceding assistant message has ToolCall (ID="call_1", Name="exec")
- **THEN** the system SHALL use "exec" as the tool_call_name for Gemini content building

#### Scenario: No matching assistant message skips tool response
- **WHEN** a tool message has metadata tool_call_id="call_x" but no tool_call_name, AND no preceding assistant message has a matching ToolCall ID
- **THEN** the tool message SHALL be skipped (not included in Gemini content)

### Requirement: OpenAI orphaned tool call repair
The system SHALL inject synthetic error tool responses in the OpenAI provider when an assistant tool call is followed by a non-tool message without a matching tool response.

#### Scenario: Orphaned tool call gets synthetic response
- **WHEN** an assistant message has ToolCall ID="call_1" followed by a user message with no intervening tool response
- **THEN** a synthetic tool message with tool_call_id="call_1" and error content SHALL be inserted before the user message

#### Scenario: Partially answered multi-call gets repair for missing only
- **WHEN** an assistant message has ToolCalls [ID="call_a", ID="call_b"] and only call_a has a tool response before the next user message
- **THEN** a synthetic tool message SHALL be inserted only for call_b

#### Scenario: Trailing pending tool call is untouched
- **WHEN** an assistant message with ToolCalls is the last message in history (no following non-tool message)
- **THEN** no synthetic tool response SHALL be inserted

### Requirement: Gemini ThoughtSignature corruption defense
The system SHALL drop replayed FunctionCall parts where `Thought=true` but `ThoughtSignature` is empty, as these represent corrupted persistence entries that Gemini API rejects.

#### Scenario: Thought=true with empty ThoughtSignature is dropped
- **WHEN** a replayed assistant ToolCall has Thought=true and ThoughtSignature=nil
- **THEN** the FunctionCall part SHALL NOT be included in Gemini content

#### Scenario: Thought=false with empty ThoughtSignature passes through
- **WHEN** a replayed assistant ToolCall has Thought=false and ThoughtSignature=nil
- **THEN** the FunctionCall part SHALL be included normally (non-thinking model)

#### Scenario: Thought=true with valid ThoughtSignature passes through
- **WHEN** a replayed assistant ToolCall has Thought=true and ThoughtSignature=[]byte("sig123")
- **THEN** the FunctionCall part SHALL be included with ThoughtSignature preserved
