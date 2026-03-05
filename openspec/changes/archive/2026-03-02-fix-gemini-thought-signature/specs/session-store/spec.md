## MODIFIED Requirements

### Requirement: ToolCall persists thinking metadata
The `session.ToolCall` and `entschema.ToolCall` structs SHALL include `Thought bool` and `ThoughtSignature []byte` fields with `omitempty` JSON tags. These fields SHALL survive the full persistence round-trip: session → database → session reload.

#### Scenario: Persist and reload ThoughtSignature
- **WHEN** a session message with FunctionCall ToolCalls containing `ThoughtSignature` is persisted via `AppendMessage`
- **THEN** retrieving the session via `Get` SHALL return ToolCalls with the original `Thought` and `ThoughtSignature` values intact

#### Scenario: Legacy session without thinking fields
- **WHEN** an existing session record has ToolCalls without `thought` or `thoughtSignature` JSON keys
- **THEN** deserialization SHALL produce `Thought=false` and `ThoughtSignature=nil` (zero values)
