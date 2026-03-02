## Why

Gemini 3+ models require `ThoughtSignature` (an opaque `[]byte`) to be preserved on `genai.Part` when sending FunctionCall parts back in conversation history. The current provider abstraction round-trip strips this field, causing `Error 400: Function call is missing a thought_signature in functionCall parts`. This blocks all tool-calling conversations with Gemini 3 models.

## What Changes

- Add `Thought bool` and `ThoughtSignature []byte` fields to ToolCall structs at all 3 layers (provider, session, ent schema)
- Capture `ThoughtSignature` and `Thought` from Gemini streaming responses
- Restore these fields when reconstructing `genai.Part` for Gemini API requests
- Propagate through session persistence (AppendEvent → EntStore → session history replay)
- Upgrade ADK Go dependency from v0.4.0 to v0.5.0

## Capabilities

### New Capabilities

_(none)_

### Modified Capabilities

- `provider-interface`: Add `Thought` and `ThoughtSignature` fields to `provider.ToolCall` struct for Gemini thinking metadata passthrough
- `session-store`: Add `Thought` and `ThoughtSignature` fields to `session.ToolCall` for persistence across session reload
- `gemini-content-sanitization`: Preserve `ThoughtSignature` through Gemini message construction and response parsing

## Impact

- **Code**: `internal/provider/provider.go`, `internal/session/store.go`, `internal/ent/schema/message.go`, `internal/provider/gemini/gemini.go`, `internal/adk/model.go`, `internal/adk/session_service.go`, `internal/adk/state.go`, `internal/session/ent_store.go`
- **Dependencies**: `google.golang.org/adk` v0.4.0 → v0.5.0
- **Backward compatibility**: Fully backward compatible — `omitempty` JSON tags ensure existing sessions deserialize cleanly; non-Gemini providers ignore zero-valued fields
