## 1. Provider Interface — StreamEventThought

- [x] 1.1 Add `StreamEventThought` constant to `internal/provider/provider.go`
- [x] 1.2 Add `ThoughtLen int` field to `StreamEvent` struct
- [x] 1.3 Update `Valid()` and `Values()` to include `StreamEventThought`

## 2. Gemini Provider — Thought Event Emission

- [x] 2.1 Modify `gemini.go` Generate to emit `StreamEventThought` with `ThoughtLen` for `Thought=true` text parts instead of silently dropping
- [x] 2.2 Preserve existing `StreamEventPlainText` emission for `Thought=false` text parts

## 3. ModelAdapter — Thought Event Handling

- [x] 3.1 Add `StreamEventThought` no-op case in streaming path of `model.go`
- [x] 3.2 Add `StreamEventThought` no-op case in non-streaming path of `model.go`

## 4. Agent — Dead Code Removal and Diagnostics

- [x] 4.1 Remove `!part.Thought` filter from `runAndCollectOnce` partial path (line ~371)
- [x] 4.2 Remove `!part.Thought` filter from `runAndCollectOnce` non-streaming path (line ~379)
- [x] 4.3 Remove `!part.Thought` filter from `RunStreaming` partial path (line ~442)
- [x] 4.4 Remove `!part.Thought` filter from `RunStreaming` non-streaming path (line ~452)
- [x] 4.5 Add warn log in `RunAndCollect` when response is empty

## 5. Empty Response Fallback

- [x] 5.1 Add `emptyResponseFallback` constant and guard in `internal/app/channels.go:runAgent`
- [x] 5.2 Add `emptyResponseFallback` constant and guard in `internal/gateway/server.go:handleChatMessage`

## 6. Verification

- [x] 6.1 Run `go build ./...` and confirm no compile errors
- [x] 6.2 Run `go test ./...` and confirm all tests pass
