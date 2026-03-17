## Context

OpenAI and Gemini produce 400 errors during session replay with tool calls. The shared `repairOrphanedFunctionCalls()` in `convertMessages()` attempted to fix orphaned tool calls for all providers, but OpenAI and Gemini have fundamentally different API contracts:

- **OpenAI**: Requires every assistant `tool_calls` entry to have a matching `tool` message with `tool_call_id`. Missing matches cause `No tool output found for function call`.
- **Gemini**: Requires `FunctionResponse.Name` to match `FunctionCall.Name`, uses `FunctionCall.ID`/`FunctionResponse.ID` for correlation, and rejects `Thought=true` parts without valid `ThoughtSignature`.

The shared repair injected synthetic tool messages without `tool_call_name` in metadata, causing Gemini to silently drop them (name was required but missing).

## Goals / Non-Goals

**Goals:**
- Fix OpenAI session replay 400 errors for orphaned tool calls
- Fix Gemini session replay errors: silent tool response drops, missing `thought_signature`
- Preserve backward compatibility with legacy sessions that lack `tool_call_name` metadata
- Reduce shared repair logic blast radius by making orphan repair provider-specific

**Non-Goals:**
- Rewriting the entire provider message conversion pipeline
- Adding retry/recovery logic for API errors
- Changing the session persistence format

## Decisions

### D1: Store `tool_call_name` alongside `tool_call_id` in metadata
**Rationale**: Gemini requires function name for `FunctionResponse`. Without it in metadata, the Gemini provider cannot reconstruct the response and silently drops it.
**Alternative**: Look up name from preceding assistant message every time → more fragile, O(n) per message.

### D2: Move orphan repair to OpenAI-specific private helper
**Rationale**: Only OpenAI requires synthetic tool responses for orphaned calls. Gemini has its own `sanitize.go` with `ensureFunctionResponsePairs()`. Keeping repair in the shared path forces all providers to handle synthetic messages they don't need.
**Alternative**: Keep shared but add provider flags → increases coupling, harder to reason about.

### D3: Backward-compatible `inferToolNameFromHistory()` for Gemini
**Rationale**: Existing sessions lack `tool_call_name`. Scanning backward for the nearest assistant message's matching ToolCall ID provides a reasonable fallback. Only checks the nearest assistant to avoid wrong inference.
**Alternative**: Require migration of all sessions → too disruptive for users.

### D4: Narrow ThoughtSignature defense (not blanket drop)
**Rationale**: Only `Thought=true && ThoughtSignature empty` is corrupted (persistence lost the signature). `Thought=false && ThoughtSignature empty` is normal for non-thinking models. Blanket filtering would break non-thinking model calls.

### D5: Set FunctionCall.ID and FunctionResponse.ID in Gemini content
**Rationale**: Gemini SDK supports ID fields for correlation. Setting them enables proper round-trip matching. ID was previously omitted, falling back to Name-only matching which is ambiguous for multiple calls to the same tool.

## Risks / Trade-offs

- **[Risk]** Legacy sessions with many orphaned calls may hit Gemini API limits after repair removal → **Mitigation**: Gemini's `sanitize.go` `ensureFunctionResponsePairs()` handles this independently
- **[Risk]** `inferToolNameFromHistory()` may not find a match if history was truncated → **Mitigation**: Falls back to skip (continue), same as current behavior for missing name
- **[Trade-off]** Provider-specific repair means duplicated logic concepts → Accepted: the implementations differ enough that sharing creates more problems than it solves
