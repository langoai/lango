## Why

Approval-gated browser actions can loop within a single agent turn because a normal one-shot approval does not protect identical retries, and approval timeout/deny results are fed back as ordinary tool failures. This causes repeated approval prompts for the same URL and burns turn budget without making progress.

## What Changes

- Add turn-local approval state for the current request so identical `tool + params` retries can reuse an earlier approval decision.
- Treat approval deny/timeout/unavailable outcomes as structured approval failures instead of generic stringly tool errors.
- Reuse one-shot `Approve` for identical retries within the same turn, while keeping `Always Allow` as the only session-wide persistent grant.
- Block duplicate approval re-prompts for identical denied/expired/unavailable calls within the same turn.
- Add structured approval observability logs that show request, callback, grant scope, and replay-block outcomes.
- Add navigator prompt guidance not to blindly retry the same browser action after approval failure.

## Capabilities

### New Capabilities

### Modified Capabilities

- `channel-approval`: Approval routing now supports turn-local replay protection and structured approval error outcomes.
- `channel-telegram`: Telegram approval flow now emits stronger approval observability signals and participates in the turn-local replay model.
- `agent-error-handling`: Approval failures are surfaced with stable non-retryable messages rather than ambiguous generic tool failure text.
- `agent-prompting`: Navigator/browser prompt guidance now discourages immediate reissue of the same browser action after approval failure.

## Impact

- `internal/approval/`
- `internal/toolchain/mw_approval.go`
- `internal/app/channels.go`
- `internal/gateway/server.go`
- `internal/channels/telegram/approval.go`
- `internal/adk/errors.go`
- `prompts/TOOL_USAGE.md`
- `prompts/agents/navigator/IDENTITY.md`
- `internal/agentregistry/defaults/navigator/AGENT.md`
