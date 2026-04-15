## 1. Interruptible Streaming (1A)

- [x] 1.1 Add `pendingRedirectInput string` field to `ChatModel` in `internal/cli/chat/chat.go`
- [x] 1.2 Change `input.go` `SetState(stateStreaming)` from `Blur()` to `Focus()` with redirect-aware placeholder
- [x] 1.3 Add `stateStreaming` to `inputAcceptsText()` in `chat.go`
- [x] 1.4 Extend `handleStreamingKey()` to handle Enter with non-empty input: store in `pendingRedirectInput`, call `cancelFn()`, finalize partial stream with `[interrupted]` marker
- [x] 1.5 Add short-circuit in `DoneMsg` handler before `stateFailed` path: if `pendingRedirectInput != ""`, skip error display, transition to `stateIdle`, call `submitCmd(pendingRedirectInput)`, clear field
- [x] 1.6 Add unit tests for redirect queue pattern: redirect during streaming, redirect with empty input, DoneMsg without redirect

## 2. Recovery Action Mapping (1B-i)

- [x] 2.1 Add `RecoveryAction` type (`Retry`, `AbortWithHint`) and `recoveryActionFor()` function in `internal/adk/errors.go`
- [x] 2.2 Add table-driven tests for `recoveryActionFor()` covering all cause classes
- [x] 2.3 Add retry loop in `Runner.Run()` (`internal/turnrunner/runner.go`): per-attempt context creation, `recoveryActionFor()` check, jittered exponential backoff, max 3 attempts
- [x] 2.4 Emit `RecoveryInfo` from Runner-level attempt loop via `traceRecorder.recordRecovery()`
- [x] 2.5 Add unit tests for retry loop: retryable error retries, non-retryable exits, max attempts exhausted, successful first attempt

## 3. Stale Stream Detection (1B-ii)

- [x] 3.1 Add `staleTimeout` field to `Runner` config with default 30s
- [x] 3.2 Add watchdog timer logic in `wrapChunkCallback()`: start on first chunk, reset on each chunk, cancel attempt context on fire
- [x] 3.3 Add unit tests for stale detection: timer reset on chunk, stale fires after timeout, timer inactive before first chunk

## 4. Inline Emergency Compaction (1B-iii)

- [x] 4.1 Define `SessionCompactor` interface in `internal/adk/context_model.go`
- [x] 4.2 Add `WithSessionCompactor()` method to `ContextAwareModelAdapter`
- [x] 4.3 Wire `EntStore` as `SessionCompactor` in `internal/app/wiring.go`
- [x] 4.4 Add emergency compaction check in `GenerateContent()` after Phase 2: trigger on `measured > modelWindow × 0.9`, NOT on `budgets.Degraded`
- [x] 4.5 Implement compaction logic: preserve first 3 + last 6 messages, summarize middle, invoke `CompactMessages()`, restart from Phase 1, at most once per call
- [x] 4.6 Add warning log + user message for `budgets.Degraded` (config issue, not compactable)
- [x] 4.7 Add unit tests for emergency compaction: trigger at 90%, no trigger below 90%, Degraded does not trigger, at-most-once guard

## 5. Verification

- [x] 5.1 Run `go build ./...` and verify no compilation errors
- [x] 5.2 Run `go test ./internal/adk/... ./internal/turnrunner/... ./internal/cli/chat/...` and verify all tests pass
- [x] 5.3 Manual test: start TUI chat, send a query, type and submit new input during streaming → verify seamless redirect (user action required)
- [x] 5.4 Manual test: verify existing Ctrl+C cancel (without redirect) still shows "Generation cancelled." (user action required)
