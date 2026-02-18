## Context

The approval system uses `CompositeProvider` to route tool approval requests based on session key prefixes (e.g., `telegram:*` → Telegram provider). Automation systems generate session keys with `cron:` or `bg:` prefixes that don't match any registered provider, causing TTY fallback failures in non-interactive environments.

However, automation systems already carry channel information: cron jobs have `DeliverTo` (e.g., `["telegram:123456789"]`) and background tasks have `OriginSession`/`OriginChannel` from the session that spawned them.

## Goals / Non-Goals

**Goals:**
- Route tool approval requests from automation systems to the originating channel
- Zero changes to existing channel approval providers or CompositeProvider
- Backward-compatible: non-automation sessions work exactly as before

**Non-Goals:**
- Workflow engine approval routing (no channel info available yet)
- Multi-channel approval (broadcasting to multiple channels)
- Approval policy configuration for automation (e.g., auto-approve certain tools)

## Decisions

**Decision 1: Context-based approval target override**

Inject an explicit approval target into the context that `wrapWithApproval` checks before falling back to the session key. This was chosen over modifying `CompositeProvider` because:
- No changes to the routing infrastructure
- Each automation system controls its own injection logic
- The override is transparent — session key still works for non-automation paths

Alternative considered: Adding `cron:`/`bg:` prefix handlers to `CompositeProvider`. Rejected because it would couple the approval system to automation internals and require parsing automation session keys to extract channel info.

**Decision 2: First DeliverTo target for cron**

Use `job.DeliverTo[0]` as the approval target. Only inject when it contains `:` (e.g., `"telegram:123456789"`) to skip bare channel names (e.g., `"telegram"`) that can't route to a specific chat.

**Decision 3: OriginSession priority for background**

Prefer `task.OriginSession` (full session key like `"telegram:123:456"`) over `task.OriginChannel` (channel identifier like `"telegram:123"`). Both work with channel providers, but OriginSession preserves the original session context.

## Risks / Trade-offs

- [Single channel approval] Cron jobs with multiple `DeliverTo` targets only route approval to the first one → Acceptable for now; multi-channel approval is a future enhancement
- [Missing channel info] If `DeliverTo` is empty or contains bare names, falls back to TTY (same as current behavior) → No regression
- [Approval timeout] If the target user is unavailable, approval will time out → Expected behavior, same as interactive sessions
