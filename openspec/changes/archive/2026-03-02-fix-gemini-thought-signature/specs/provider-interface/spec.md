## MODIFIED Requirements

### Requirement: ToolCall carries provider-specific metadata
The `provider.ToolCall` struct SHALL include `Thought bool` and `ThoughtSignature []byte` fields to support Gemini thinking metadata passthrough. These fields SHALL be zero-valued for non-Gemini providers.

#### Scenario: Gemini FunctionCall with ThoughtSignature
- **WHEN** a Gemini streaming response contains a FunctionCall part with `Thought=true` and `ThoughtSignature` set
- **THEN** the resulting `provider.ToolCall` SHALL have `Thought=true` and `ThoughtSignature` populated with the original bytes

#### Scenario: Non-Gemini provider ToolCall
- **WHEN** a non-Gemini provider emits a ToolCall
- **THEN** the `Thought` field SHALL be `false` and `ThoughtSignature` SHALL be `nil`
