## Why

OpenAI API returns 400 error (`Invalid 'input[12].name': empty string`) when replaying session history. The root cause is that streaming tool call deltas (which carry only partial arguments, no Name) are stored as individual FunctionCall parts with empty names. On the next turn, these empty-name parts are sent back to OpenAI, which rejects them.

## What Changes

- Add `Index *int` field to `provider.ToolCall` for OpenAI streaming chunk correlation
- Introduce `toolCallAccumulator` in `internal/adk/model.go` that assembles streaming deltas into complete tool calls using a fallback chain: `Index` (OpenAI) → `ID`/`Name` (Anthropic start) → last active entry (Anthropic delta)
- Refactor streaming and non-streaming paths in `ModelAdapter.GenerateContent` to use the accumulator instead of storing each delta as a separate FunctionCall part
- Add defensive filters in `convertMessages()`, `convertTools()`, and OpenAI `convertParams()` to skip entries with empty names

## Capabilities

### New Capabilities

- `streaming-tool-call-assembly`: Accumulator-based assembly of streaming tool call deltas into complete FunctionCall parts, supporting both OpenAI (Index-based) and Anthropic (ID/Name-based) streaming patterns

### Modified Capabilities

- `provider-openai-compatible`: Add Index pass-through and empty-name safety filters in convertParams

## Impact

- `internal/provider/provider.go` — ToolCall struct gains Index field
- `internal/provider/openai/openai.go` — Index forwarding + convertParams safety filters
- `internal/adk/model.go` — toolCallAccumulator type + streaming refactor + convertMessages/convertTools guards
- No breaking changes to external APIs; purely internal fix
