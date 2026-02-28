## 1. Slack Mock Thread Safety

- [x] 1.1 Add `sync.Mutex` field to `MockClient` struct in `internal/channels/slack/slack_test.go`
- [x] 1.2 Add mutex Lock/Unlock in `PostMessage()` and `UpdateMessage()` around slice appends
- [x] 1.3 Add `getPostMessages()` and `getUpdateMessages()` helper methods returning defensive copies
- [x] 1.4 Replace direct field access in `TestSlackChannel` and `TestSlackThinkingPlaceholder` with helper methods

## 2. Telegram Mock Thread Safety

- [x] 2.1 Add `sync.Mutex` field to `MockBotAPI` struct in `internal/channels/telegram/telegram_test.go`
- [x] 2.2 Add mutex Lock/Unlock in `Send()` and `Request()` around slice appends
- [x] 2.3 Add `getSentMessages()` and `getRequestCalls()` helper methods returning defensive copies
- [x] 2.4 Replace direct field access in `TestTelegramChannel` and `TestTelegramTypingIndicator` with helper methods

## 3. Background Process Output Thread Safety

- [x] 3.1 Add `syncBuffer` type to `internal/tools/exec/exec.go` wrapping `bytes.Buffer` with `sync.Mutex`
- [x] 3.2 Implement `Write(p []byte) (int, error)` and `String() string` on `syncBuffer`
- [x] 3.3 Change `BackgroundProcess.Output` type from `*bytes.Buffer` to `*syncBuffer`
- [x] 3.4 Update `StartBackground()` to create `*syncBuffer` instead of `*bytes.Buffer`

## 4. Verification

- [x] 4.1 Run `go test -race ./internal/channels/slack/...` — pass with no races
- [x] 4.2 Run `go test -race ./internal/channels/telegram/...` — pass with no races
- [x] 4.3 Run `go test -race ./internal/supervisor/...` — pass with no races
- [x] 4.4 Run `go build ./...` and `go test ./...` — all pass
