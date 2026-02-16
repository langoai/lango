## Context

The approval routing system relies on the session key (e.g., `"telegram:123:456"`) being available in `context.Context` so that `CompositeProvider` can match it against registered channel providers via `CanHandle()`. Currently, `runAgent()` in `channels.go` receives the session key as a parameter but does not inject it into the context. This causes `SessionKeyFromContext(ctx)` to return an empty string, making all channel providers return `false` from `CanHandle("")`. The request then falls through to TTY fallback (which denies in non-terminal environments) or the fail-closed path (which silently returns `false, nil`).

## Goals / Non-Goals

**Goals:**
- Ensure session key flows through context from channel handler to approval provider
- Make fail-closed denial observable via error return (not silent `nil`)
- Help the AI distinguish temporary denial from permanent restriction via error messages
- Add system prompt guidance so the AI correctly handles approval outcomes

**Non-Goals:**
- Changing the approval flow architecture or provider interface
- Adding retry logic or automatic re-approval
- Modifying channel-specific provider implementations (Telegram, Discord, Slack)

## Decisions

### Decision 1: Inject session key in runAgent, not in each channel handler
The `WithSessionKey` call is placed in the single `runAgent()` function rather than in each of the three channel handler methods. This is the last common point before the agent pipeline and avoids duplication. The `WithSessionKey` function already exists in `tools.go`.

### Decision 2: Sentinel error from CompositeProvider fail-closed path
Changed `return false, nil` to `return false, fmt.Errorf(...)` when no provider matches and no TTY fallback exists. This makes the failure observable in logs and allows `wrapWithApproval` to propagate a meaningful error. The error includes the session key for debugging.

### Decision 3: Differentiated error messages in wrapWithApproval
When approval is denied, the error message now checks whether a session key was present. This gives the AI two distinct signals:
- "no approval channel available (session key missing)" → system configuration problem
- "user did not approve the action" → user chose to deny

### Decision 4: System prompt guidance over code-level retry
Rather than adding automatic retry or complex error recovery in code, we add a "Tool Approval" section to `TOOL_USAGE.md`. This lets the AI interpret denial errors correctly and communicate with the user, which is the appropriate response for an interactive approval flow.

## Risks / Trade-offs

- [Breaking change for code testing `err == nil` on fail-closed] → Only affects `TestCompositeProvider_FailClosed`, which is updated in the same change. External consumers relying on `(false, nil)` will now see an error, which is the correct behavior for fail-closed.
- [AI may still misinterpret errors if prompt not loaded] → Mitigated by the error messages being self-explanatory even without the prompt context.
