## Why

The `runAgent()` session key injection fix (previous change) revealed an identical anti-pattern across the codebase: `context.Background()` used where a parent context should propagate, and errors silently discarded as `nil`. These cause approval routing to fail (no session key), cancellation signals to be lost, and debugging to be difficult when audit log writes fail silently.

## What Changes

- Move `WithSessionKey`/`SessionKeyFromContext` from `internal/app/tools.go` to shared `internal/session/context.go` so both `app` and `gateway` packages can use them
- Replace `context.Background()` in Discord's `onMessageCreate` callback with the `Start(ctx)` context
- Inject session key into context in Gateway's `handleChatMessage` via `session.WithSessionKey`
- Return an error from `TTYProvider.RequestApproval` when stdin is not a terminal (instead of silent deny)
- Add comma-ok guards on all `sync.Map` type assertions in Discord, Telegram, and Slack approval providers
- Log audit log save errors instead of discarding them with `_ =`
- Add a `default` case in Anthropic provider's message role switch to warn on unknown roles

## Capabilities

### New Capabilities

- `session-context-helpers`: Shared session key context injection/extraction utilities in `internal/session`

### Modified Capabilities

- `channel-approval`: Safe type assertions in approval providers; TTY now returns error on non-terminal
- `channel-discord`: Context propagation from `Start(ctx)` to message handler callback
- `gateway-server`: Session key injection into context for downstream approval routing
- `provider-anthropic`: Unknown role handling with warning log

## Impact

- `internal/session/context.go` (new file)
- `internal/app/tools.go` — session helpers removed, audit log error handling added
- `internal/app/channels.go` — updated import path
- `internal/channels/discord/discord.go` — ctx field added, Start context propagated
- `internal/channels/discord/approval.go` — safe type assertion
- `internal/channels/telegram/approval.go` — safe type assertion
- `internal/channels/slack/approval.go` — safe type assertion
- `internal/gateway/server.go` — session key injected into context
- `internal/approval/tty.go` — error returned on non-terminal
- `internal/approval/tty_test.go` (new file)
- `internal/provider/anthropic/anthropic.go` — default case + logger
