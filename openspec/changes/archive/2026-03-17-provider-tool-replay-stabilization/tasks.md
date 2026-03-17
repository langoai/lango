## 1. Shared Layer (internal/adk/model.go)

- [x] 1.1 Add `tool_call_name` to FunctionResponse metadata in `convertMessages()`
- [x] 1.2 Remove `repairOrphanedFunctionCalls()` call from `convertMessages()`
- [x] 1.3 Remove `repairOrphanedFunctionCalls()` function definition
- [x] 1.4 Update model_test.go: change orphan test to verify no-repair behavior, add tool_call_name assertion, remove partial response test

## 2. Gemini Provider (internal/provider/gemini/gemini.go)

- [x] 2.1 Set `FunctionCall.ID` from ToolCall.ID in assistant message content builder
- [x] 2.2 Set `FunctionResponse.ID` from metadata `tool_call_id` in tool message content builder
- [x] 2.3 Use `FunctionCall.ID` with Name fallback in streaming response handler
- [x] 2.4 Add `inferToolNameFromHistory()` backward-compat helper for legacy sessions missing `tool_call_name`
- [x] 2.5 Wire `inferToolNameFromHistory()` into tool message handling with fallback logic
- [x] 2.6 Add ThoughtSignature corruption defense: drop `Thought=true && ThoughtSignature empty` FunctionCalls
- [x] 2.7 Add logger variable for warn-level logging

## 3. OpenAI Provider (internal/provider/openai/openai.go)

- [x] 3.1 Add `repairOrphanedToolCalls()` private helper (same logic as removed shared function)
- [x] 3.2 Wire `repairOrphanedToolCalls()` into `convertParams()` before message conversion

## 4. Tests

- [x] 4.1 Add gemini_test.go: `TestInferToolNameFromHistory` table-driven tests (match, no-match, nearest-only)
- [x] 4.2 Add gemini_test.go: `TestThoughtSignatureFiltering` table-driven tests (3 conditions)
- [x] 4.3 Add gemini_test.go: `TestFunctionCallID_InContent` verifying ID propagation
- [x] 4.4 Add gemini_test.go: `TestInferToolNameFromHistory_BackwardCompat` legacy session test
- [x] 4.5 Add openai_test.go: `TestRepairOrphanedToolCalls` for orphan, partial, trailing, matched cases
- [x] 4.6 Add openai_test.go: `TestConvertParams_RepairIntegration` verifying repair wiring
- [x] 4.7 Add gemini_test.go: `TestStreamingFunctionCallID` for streaming ID preference/fallback scenarios

## 5. Verification

- [x] 5.1 `go build ./...` succeeds
- [x] 5.2 `go test ./internal/adk/ ./internal/provider/openai/ ./internal/provider/gemini/` all pass
- [x] 5.3 `go test ./...` full suite passes
