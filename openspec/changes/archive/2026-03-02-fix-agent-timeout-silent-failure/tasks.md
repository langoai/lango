## 1. ADK Context Deadline Detection

- [x] 1.1 Add `ctx.Err()` check after iterator loop in `runAndCollectOnce` (`internal/adk/agent.go`)
- [x] 1.2 Add `ctx.Err()` check after iterator loop in `RunStreaming` (`internal/adk/agent.go`)
- [x] 1.3 Add tests for context cancellation and deadline exceeded detection (`internal/adk/agent_test.go`)

## 2. Gateway Error Event Handling

- [x] 2.1 Replace unconditional `agent.done` with error-aware branching in `handleChatMessage` (`internal/gateway/server.go`): broadcast `agent.error` on failure, `agent.done` on success only
- [x] 2.2 Add 80% timeout warning timer that broadcasts `agent.warning` event (`internal/gateway/server.go`)
- [x] 2.3 Add tests for `agent.error` event on failure, `agent.done` on success, and nil-agent early return (`internal/gateway/server_test.go`)

## 3. Channel Typing Indicator Safety

- [x] 3.1 Add `sync.Once` to private `startTyping` stop function in Discord (`internal/channels/discord/discord.go`)
- [x] 3.2 Add `sync.Once` to private `startTyping` stop function in Telegram (`internal/channels/telegram/telegram.go`)

## 4. Verification

- [x] 4.1 Run `go build ./...` and confirm no errors
- [x] 4.2 Run `go test ./...` and confirm all tests pass
