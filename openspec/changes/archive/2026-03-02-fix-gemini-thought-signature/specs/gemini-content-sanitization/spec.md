## MODIFIED Requirements

### Requirement: Gemini message builder preserves ThoughtSignature
When constructing `genai.Content` for Gemini API requests, the message builder SHALL set `Thought` and `ThoughtSignature` on `genai.Part` from the corresponding `provider.ToolCall` fields.

#### Scenario: Reconstruct FunctionCall with ThoughtSignature for API request
- **WHEN** a `provider.Message` with ToolCalls containing `ThoughtSignature` is converted to `genai.Content`
- **THEN** the resulting `genai.Part` SHALL have `Thought` and `ThoughtSignature` fields set to match the ToolCall values

#### Scenario: Session history replay preserves ThoughtSignature
- **WHEN** session history is converted to ADK events via `EventsAdapter`
- **THEN** FunctionCall `genai.Part` instances SHALL include `Thought` and `ThoughtSignature` from the stored `session.ToolCall`

### Requirement: ModelAdapter propagates ThoughtSignature bidirectionally
The `ModelAdapter` SHALL propagate `Thought` and `ThoughtSignature` from `provider.ToolCall` to `genai.Part` in both streaming and non-streaming paths, and from `genai.Part` to `provider.ToolCall` in `convertMessages`.

#### Scenario: Streaming path preserves ThoughtSignature
- **WHEN** a streaming ToolCall event arrives with `ThoughtSignature`
- **THEN** the `genai.Part` yielded in `LLMResponse` SHALL include the `Thought` and `ThoughtSignature` fields

#### Scenario: convertMessages extracts ThoughtSignature
- **WHEN** `convertMessages` processes a `genai.Content` with FunctionCall parts carrying `ThoughtSignature`
- **THEN** the resulting `provider.ToolCall` SHALL include the `Thought` and `ThoughtSignature` values
