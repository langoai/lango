## Why

Gemini API rejects requests with "Please ensure that function call turn comes immediately after a user turn or after a function response turn" when tool call history is loaded from DB. This occurs because `FunctionResponse` structure (Name, ID) is lost during the save/restore cycle, causing ADK to misinterpret tool events as foreign agent messages and breaking the required message ordering.

## What Changes

- Preserve original `FunctionCall.ID` during `AppendEvent` instead of replacing with synthetic `"call_" + name`
- Store `FunctionResponse` metadata (ID, Name, Response) in `ToolCalls` JSON column alongside the response content
- Reconstruct `genai.FunctionResponse` parts when loading history in `EventsAdapter.All()`, with legacy fallback for existing sessions
- Add sequence safety to `tokenBudgetTruncate` to prevent truncated history from starting with orphaned tool/function messages
- Forward original `FunctionCall.ID` in `convertMessages` for provider-level call-response matching

## Capabilities

### New Capabilities

### Modified Capabilities
- `adk-architecture`: Fix FunctionCall/FunctionResponse save/restore cycle to preserve IDs and reconstruct proper genai parts
- `session-store`: Extend ToolCalls usage to store FunctionResponse metadata (no schema change, data-only addition to JSON column)

## Impact

- **Files modified**: `internal/adk/session_service.go`, `internal/adk/state.go`, `internal/adk/model.go`
- **Tests added**: `internal/adk/state_test.go`, `internal/adk/session_service_test.go`
- **DB schema**: No changes (uses existing `tool_calls` JSON column)
- **Backward compatibility**: Existing sessions without FunctionResponse metadata use position-based legacy fallback
