## Why

When the ADK library stores FunctionResponse events, it sets `Content.Role = "user"` instead of `"tool"` or `"function"`. During session reconstruction via `EventsAdapter.All()`, messages with role `"user"` are treated as plain text, causing FunctionResponse data to be silently dropped. On retry, OpenAI detects an orphaned FunctionCall (no matching tool response) and returns a 400 error: `"No tool output found for function call"`.

## What Changes

- **Write-time role correction** (`session_service.go`): `AppendEvent` detects FunctionResponse-only messages arriving with role `"user"` and corrects to `"tool"` before persisting.
- **Read-time legacy data correction** (`state.go`): `EventsAdapter.All()` corrects role for FunctionResponse messages already stored with `"user"` in existing databases.
- **Provider boundary defense** (`model.go`): `repairOrphanedFunctionCalls` injects synthetic error responses for orphaned FunctionCalls that have a following user message but no intervening tool response. Pending calls at the end of history are never touched.

## Capabilities

### New Capabilities

_(none — this is a bug fix, not a new capability)_

### Modified Capabilities

- `adk-architecture`: Session event storage and reconstruction must preserve correct roles for FunctionResponse events, and the provider message conversion must handle orphaned FunctionCalls defensively.

## Impact

- `internal/adk/session_service.go` — AppendEvent write path
- `internal/adk/state.go` — EventsAdapter.All() read path
- `internal/adk/model.go` — convertMessages provider boundary
- Existing databases with incorrectly stored FunctionResponse events are automatically corrected at read-time (no migration needed)
