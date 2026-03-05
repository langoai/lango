## Why

When an agent request times out, no error message is delivered to the user and typing/thinking indicators are never terminated. The root cause is that the Google ADK streaming iterator silently terminates on context deadline exceeded without yielding an error, causing `RunAndCollect`/`RunStreaming` to return `("", nil)` — making the timeout completely invisible to all downstream handlers (channels and gateway).

## What Changes

- Add post-iteration `ctx.Err()` check in `runAndCollectOnce` and `RunStreaming` to detect context deadline exceeded that ADK's iterator fails to propagate
- Replace unconditional `agent.done` broadcast in Gateway with error-aware branching: `agent.error` on failure, `agent.done` on success only
- Add `agent.warning` event broadcast at 80% timeout in Gateway for proactive user notification
- Add `sync.Once` safety to private `startTyping` stop functions in Discord and Telegram channels

## Capabilities

### New Capabilities

_(none — all changes modify existing capabilities)_

### Modified Capabilities
- `agent-runtime`: Add context deadline detection after ADK iterator completion to surface silent timeouts as errors
- `gateway-server`: Introduce `agent.error` and `agent.warning` WebSocket events; `agent.done` sent only on success
- `thinking-indicator`: Add `sync.Once` double-close safety to private `startTyping` functions in Discord and Telegram

## Impact

- `internal/adk/agent.go` — `runAndCollectOnce`, `RunStreaming` functions
- `internal/gateway/server.go` — `handleChatMessage` function
- `internal/channels/discord/discord.go` — `startTyping` function
- `internal/channels/telegram/telegram.go` — `startTyping` function
- WebSocket clients must handle new `agent.error` and `agent.warning` events (backward compatible — unknown events are ignored)
