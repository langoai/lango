## Why

Automation systems (cron, background) fail when tools require approval because the session keys they generate (`cron:JobName:timestamp`, `bg:taskID`) don't match any registered approval provider, causing a TTY fallback that fails in non-interactive environments ("stdin is not a terminal"). These systems already have channel routing information (cron `DeliverTo`, background `OriginSession`) that can be used to route approval requests to the originating channel.

## What Changes

- Add context-based approval target override mechanism that allows automation systems to inject the originating channel as the approval routing target
- Modify `wrapWithApproval` to check for an explicit approval target before falling back to the session key
- Inject approval targets in cron executor using `job.DeliverTo[0]` when it contains a channel identifier
- Inject approval targets in background manager using `task.OriginSession` or `task.OriginChannel`

## Capabilities

### New Capabilities

- `automation-approval-routing`: Context-based approval target override for routing tool approval requests from automation systems to the originating channel

### Modified Capabilities

- `channel-approval`: The composite routing now supports an explicit approval target that overrides the session key for provider matching

## Impact

- `internal/approval/context.go` — New file with context helpers
- `internal/app/tools.go` — `wrapWithApproval` uses approval target override
- `internal/cron/executor.go` — Injects approval target from `DeliverTo`
- `internal/background/manager.go` — Injects approval target from origin info
- No changes to channel-specific approval providers (Telegram, Discord, Slack)
- No changes to `CompositeProvider` routing logic
