## Context

After fixing the `runAgent()` session key injection bug, a codebase audit revealed the same anti-pattern in multiple locations: `context.Background()` replacing parent contexts, and errors silently discarded. The session key context helpers (`WithSessionKey`/`SessionKeyFromContext`) live in `internal/app/tools.go`, making them inaccessible to `internal/gateway` due to Go's import cycle rules. Discord's callback handler creates a fresh `context.Background()` that drops any parent cancellation/deadline signals. Gateway's `handleChatMessage` never injects the session key, so downstream approval routing fails. The TTY provider silently returns `(false, nil)` when no terminal is attached, making it indistinguishable from an explicit user denial.

## Goals / Non-Goals

**Goals:**
- Move session key context helpers to a shared package accessible by both `app` and `gateway`
- Propagate parent context through Discord channel callbacks instead of using `context.Background()`
- Inject session key into Gateway's chat message context for approval routing
- Make TTY provider return an explicit error when stdin is not a terminal
- Add safe type assertions for all `sync.Map` loads in channel approval providers
- Log audit log errors instead of discarding them

**Non-Goals:**
- Refactoring the approval provider interface
- Adding new context values beyond the existing session key
- Changing the Gateway's authentication flow

## Decisions

**1. Session helpers in `internal/session` package**
The `internal/session` package already exists (store types), making it a natural home for session-scoped context helpers. Alternative: a new `internal/ctxutil` package — rejected because session key is semantically tied to sessions.

**2. Store `Start(ctx)` in Discord Channel struct**
The discordgo library callbacks (`func(*discordgo.Session, *discordgo.MessageCreate)`) don't accept a `context.Context`. Storing the `Start(ctx)` context in the struct and using it in callbacks is the simplest approach. Alternative: creating a per-message derived context — unnecessary since the parent context already carries the right lifecycle.

**3. TTY returns error instead of silent `nil`**
`(false, nil)` is indistinguishable from user denial. Returning an error lets `CompositeProvider` distinguish "no terminal" from "user said no" and potentially fall through to another provider. This aligns with Go error handling conventions: handle errors once, don't hide them.

**4. Comma-ok on sync.Map assertions**
While the stored type is always `chan bool` or `*approvalPending` today, defensive coding prevents panics from future changes or concurrent corruption. Log + return instead of panic.

## Risks / Trade-offs

- **Discord ctx lifetime**: Storing `Start(ctx)` means all message handlers share the same context. If `Start`'s parent is cancelled, all in-flight handlers are cancelled. This is the desired behavior (graceful shutdown), but long-running handlers should be aware. → Acceptable because handler calls are relatively short-lived.
- **TTY error propagation**: Existing callers that expect `(false, nil)` for non-terminal may need adjustment. → The `CompositeProvider` already wraps this and handles errors appropriately.
