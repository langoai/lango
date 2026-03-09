## Why

When the agent performs deep research (>5 minutes), the hard 5-minute timeout discards all accumulated work and returns a generic `"request timed out after 5m0s"` error. Users see no partial results, no progress indication, and no actionable guidance. Subsystem timeouts cascade with `context deadline exceeded` errors. This degrades trust and wastes compute.

## What Changes

- Add structured `AgentError` type with error codes, partial result preservation, and user-facing hints
- Modify agent run methods to return accumulated text on failure instead of discarding it
- Add user-friendly error formatting across all channels (Slack, Telegram, Discord, Gateway)
- Replace pure typing indicators with progressive "Thinking... (30s)" messages that show elapsed time
- Add auto-extend timeout capability that extends the deadline when agent activity is detected
- Add `AutoExtendTimeout` and `MaxRequestTimeout` config fields

## Capabilities

### New Capabilities
- `agent-error-handling`: Structured error types with classification, partial result recovery, and user-facing messages
- `progress-indicators`: Progressive thinking indicators with elapsed time across all channels and gateway
- `auto-extend-timeout`: Configurable automatic deadline extension based on agent activity detection

### Modified Capabilities

## Impact

- `internal/adk/` — New `AgentError` type, modified `RunAndCollect`, `RunStreaming`, `runAndCollectOnce`
- `internal/app/` — New error formatting, `ExtendableDeadline`, modified `runAgent()`
- `internal/channels/slack/` — Progress updates on placeholder message
- `internal/channels/telegram/` — Thinking placeholder message with periodic edit
- `internal/channels/discord/` — Thinking placeholder message with periodic edit
- `internal/gateway/` — Structured error fields in `agent.error` event, `agent.progress` broadcast
- `internal/config/` — New `AutoExtendTimeout`, `MaxRequestTimeout` fields in `AgentConfig`
- No external API or dependency changes
