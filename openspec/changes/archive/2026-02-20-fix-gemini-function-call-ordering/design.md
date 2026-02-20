## Context

The ADK session adapter layer (`internal/adk/`) bridges Lango's internal session store with Google's ADK runner. When the ADK runner invokes tools, it produces events containing `genai.FunctionCall` (with ID, Name, Args) and `genai.FunctionResponse` (with ID, Name, Response). These events are persisted to the database via `AppendEvent` and restored via `EventsAdapter.All()`.

Currently, two critical data losses occur during this cycle:
1. `FunctionCall.ID` is replaced with a synthetic `"call_" + name`, breaking ADK's ID-based call-response matching
2. `FunctionResponse` metadata (Name, ID) is discarded entirely, with only the response body stored as plain text content

This causes the ADK's `rearrangeEventsForFunctionResponsesInHistory` to fail call-response matching, leading it to treat tool events as "foreign agent" messages. The result is `user(text)` messages where `function(FunctionResponse)` messages are expected, violating Gemini's turn ordering rules.

## Goals / Non-Goals

**Goals:**
- Preserve original `FunctionCall.ID` through the save/restore cycle
- Store and reconstruct `FunctionResponse` parts with full metadata (ID, Name, Response)
- Maintain backward compatibility with existing sessions that lack FunctionResponse metadata
- Prevent truncated history from starting with orphaned tool messages

**Non-Goals:**
- Changing the database schema (the existing `tool_calls` JSON column is sufficient)
- Modifying the ADK library itself
- Handling multi-turn parallel function calls (single call-response pairs per message)

## Decisions

### 1. Store FunctionResponse metadata in ToolCalls array
**Decision**: Reuse the existing `ToolCall` struct's `Output` field to store FunctionResponse data alongside `ID` and `Name`.
**Rationale**: The `tool_calls` column is already a JSON array, and `ToolCall` already has an `Output` field. This avoids schema migration while storing all needed metadata. Alternative (new DB column) was rejected due to migration complexity for zero gain.

### 2. Position-based legacy fallback for old sessions
**Decision**: When a tool message has no `ToolCalls` metadata, infer `FunctionResponse` from the preceding assistant message's `ToolCalls` using positional matching.
**Rationale**: Existing sessions stored before this fix lack FunctionResponse metadata. Rather than requiring a data migration, the adapter infers the missing data from context. Over time, token budget truncation naturally phases out legacy data.

### 3. Sequence safety only on truncation
**Decision**: Skip leading tool/function messages only when `tokenBudgetTruncate` actually removed messages (startIdx > 0).
**Rationale**: A trailing FunctionCall without a response is valid during active tool execution (response is pending). The safety check should only apply at truncation boundaries where the context has been cut.

## Risks / Trade-offs

- **Legacy fallback accuracy**: Position-based matching may misattribute responses if multiple tool calls were made in a single turn and not all responses are present. Mitigation: this matches the previous broken behavior, so it's no worse, and new sessions store explicit metadata.
- **Content duplication**: FunctionResponse data is stored in both `ToolCalls[].Output` and `msg.Content` for backward compat. Mitigation: minor storage overhead, ensures old code paths that read `Content` still work.
