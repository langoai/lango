## 1. ADK Upgrade

- [x] 1.1 Upgrade ADK v0.5.0 → v0.6.0 in go.mod (`go get google.golang.org/adk@v0.6.0`)
- [x] 1.2 Run `go mod tidy` and verify `go build ./...` passes
- [x] 1.3 Run `go test ./internal/adk/... -count=1` to confirm ADK compatibility

## 2. OpenAI Thought Call Filtering

- [x] 2.1 Add `droppedThoughtIDs` set construction in `convertParams()` pre-scan
- [x] 2.2 Filter thought tool calls from assistant messages using the set
- [x] 2.3 Drop tool response messages whose `tool_call_id` is in the dropped set
- [x] 2.4 Add test: mixed thought + normal calls filtered correctly
- [x] 2.5 Add test: all thought calls dropped leaves assistant with no tool calls
- [x] 2.6 Add test: no thought calls leaves messages unchanged

## 3. Gemini Orphaned FunctionResponse Removal

- [x] 3.1 Implement `dropOrphanedFunctionResponses()` in `sanitize.go`
- [x] 3.2 Call `dropOrphanedFunctionResponses()` after `sanitizeContents()` in `gemini.go`
- [x] 3.3 Add test: orphan response removed when FunctionCall dropped
- [x] 3.4 Add test: all FunctionCalls present — no orphans removed
- [x] 3.5 Add test: content block with only orphaned responses removed entirely

## 4. Error Classification

- [x] 4.1 Add thought_signature / thoughtSignature pattern to `classifyError()` returning `ErrModelError`
- [x] 4.2 Add test: thought_signature error classified as model error
- [x] 4.3 Add test: thoughtSignature camelCase error classified as model error

## 5. Verification

- [x] 5.1 Run `go build ./...` — full project builds
- [x] 5.2 Run `go test ./internal/provider/openai/... -count=1` — all pass
- [x] 5.3 Run `go test ./internal/provider/gemini/... -count=1` — all pass
- [x] 5.4 Run `go test ./internal/adk/... -count=1` — all pass
