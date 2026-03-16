## Why

Session replay with tool calls produces 400 errors on OpenAI (`No tool output found for function call`) and Gemini (`missing thought_signature`, silent tool response drops). The root cause is that the shared `convertMessages()` repair logic ignores provider-specific API contract differences, creating a blast radius that breaks both providers differently.

## What Changes

- Store `tool_call_name` in FunctionResponse metadata alongside `tool_call_id` so Gemini tool responses are no longer silently dropped
- Set `FunctionCall.ID` and `FunctionResponse.ID` in Gemini content builder (previously omitted)
- Use `FunctionCall.ID` with Name fallback in Gemini streaming responses
- Add backward-compatible `inferToolNameFromHistory()` for legacy sessions missing `tool_call_name`
- Move `repairOrphanedFunctionCalls` from shared `convertMessages()` to OpenAI-specific `repairOrphanedToolCalls()` private helper
- Add narrow defense for corrupted thinking entries (`Thought=true && ThoughtSignature empty`) in Gemini content builder

## Capabilities

### New Capabilities
- `provider-tool-replay`: Provider-specific tool call replay stabilization covering metadata preservation, ID propagation, backward compatibility, and corrupted entry defense

### Modified Capabilities
- `streaming-tool-call-assembly`: Remove shared `repairOrphanedFunctionCalls` call from `convertMessages()`, as orphan repair is now provider-specific

## Impact

- `internal/adk/model.go`: `tool_call_name` metadata addition, `repairOrphanedFunctionCalls` removal
- `internal/provider/gemini/gemini.go`: FunctionCall.ID/FunctionResponse.ID, streaming ID, tool_call_name fallback, ThoughtSignature defense
- `internal/provider/openai/openai.go`: `repairOrphanedToolCalls()` private helper, `convertParams()` integration
- Test files updated across all three packages
