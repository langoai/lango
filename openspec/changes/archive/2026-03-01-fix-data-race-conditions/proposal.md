## Why

CI tests fail with `-race` flag due to unsynchronized concurrent access to shared slices and buffers in 3 packages: slack channel mock, telegram channel mock, and exec background process output buffer.

## What Changes

- Add `sync.Mutex` to `MockClient` in slack tests to protect `PostMessages`/`UpdateMessages` slice access
- Add `sync.Mutex` to `MockBotAPI` in telegram tests to protect `SentMessages`/`RequestCalls` slice access
- Introduce `syncBuffer` type in `internal/tools/exec/exec.go` wrapping `bytes.Buffer` with `sync.Mutex` for thread-safe background process output
- Add thread-safe helper methods (`getPostMessages`, `getUpdateMessages`, `getSentMessages`, `getRequestCalls`) to mock types

## Capabilities

### New Capabilities

### Modified Capabilities

- `tool-exec`: `BackgroundProcess.Output` type changes from `*bytes.Buffer` to `*syncBuffer` for thread-safe concurrent read/write
- `test-coverage`: Add mutex synchronization to channel mock types to eliminate data races under `-race` flag

## Impact

- `internal/channels/slack/slack_test.go` — test-only mock changes
- `internal/channels/telegram/telegram_test.go` — test-only mock changes
- `internal/tools/exec/exec.go` — production code: new `syncBuffer` type, `BackgroundProcess.Output` type change
- `internal/supervisor/supervisor.go` — no code change needed (uses `Output.String()` which is now thread-safe)
