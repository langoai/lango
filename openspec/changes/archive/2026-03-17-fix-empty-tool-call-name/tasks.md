## 1. Provider ToolCall Index Field

- [x] 1.1 Add `Index *int` field to `provider.ToolCall` struct in `internal/provider/provider.go`
- [x] 1.2 Forward `tc.Index` from OpenAI SDK `ToolCall` to `provider.ToolCall.Index` in `internal/provider/openai/openai.go` streaming path

## 2. Tool Call Accumulator

- [x] 2.1 Implement `accumEntry` struct and `toolCallAccumulator` type in `internal/adk/model.go` with `add()` and `done()` methods
- [x] 2.2 Implement fallback chain in `add()`: Index → ID/Name → lastIndex
- [x] 2.3 Implement `done()`: sort by index, drop empty-name entries, generate fallback IDs

## 3. Streaming Path Refactor

- [x] 3.1 Replace `var toolParts []*genai.Part` with `var toolAccum toolCallAccumulator` in streaming path
- [x] 3.2 Yield partial tool call only when delta carries non-empty Name
- [x] 3.3 Use `toolAccum.done()` in StreamEventDone handler for final assembled parts

## 4. Non-Streaming Path Refactor

- [x] 4.1 Replace `var toolParts []*genai.Part` with `var toolAccum toolCallAccumulator` in non-streaming path
- [x] 4.2 Use `toolAccum.done()` for final part assembly

## 5. Defensive Filters

- [x] 5.1 Add empty-name FunctionCall skip in `convertMessages()` with warning log
- [x] 5.2 Add empty-name FunctionDeclaration skip in `convertTools()` with warning log
- [x] 5.3 Add empty-name tool filter in OpenAI `convertParams()` tools section
- [x] 5.4 Add empty-name tool call filter in OpenAI `convertParams()` messages section

## 6. Tests

- [x] 6.1 Add `TestToolCallAccumulator_SingleComplete` test
- [x] 6.2 Add `TestToolCallAccumulator_OpenAIStreaming` test
- [x] 6.3 Add `TestToolCallAccumulator_OpenAIMultipleCalls` test
- [x] 6.4 Add `TestToolCallAccumulator_AnthropicStreaming` test
- [x] 6.5 Add `TestToolCallAccumulator_AnthropicMultipleCalls` test
- [x] 6.6 Add `TestToolCallAccumulator_OrphanDeltaDropped` test
- [x] 6.7 Add `TestToolCallAccumulator_EmptyNameDropped` test
- [x] 6.8 Add `TestToolCallAccumulator_IDPreserved` test
- [x] 6.9 Add `TestGenerateContent_StreamingToolCallRegression` E2E test
- [x] 6.10 Add `TestConvertMessages_EmptyFunctionCallName` test
- [x] 6.11 Add `TestConvertTools_EmptyName` test
- [x] 6.12 Add `TestConvertParams_EmptyToolNameFiltered` test
- [x] 6.13 Add `TestConvertParams_EmptyToolCallNameFiltered` test
- [x] 6.14 Add `TestConvertParams_ValidToolsUnchanged` test

## 7. Verification

- [x] 7.1 Run `go build ./...` — full project builds without errors
- [x] 7.2 Run `go test ./internal/adk/... ./internal/provider/...` — all tests pass
- [x] 7.3 Run `go test ./...` — full test suite passes
