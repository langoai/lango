## Context

Two runtime bugs surfaced in Docker sidecar deployment mode:

1. **Browser panic**: `ensureBrowser()` has no synchronization. Concurrent browser tool calls (navigate, action, screenshot) race on `t.browser`. If `Connect()` fails, `t.browser` remains non-nil but disconnected, causing nil pointer dereference on the next `Page()` call.

2. **Unknown agent "model"**: The ADK `ModelAdapter` generates responses with `role: "model"`. When `EventsAdapter.All()` reconstructs events from stored messages with empty `Author` fields, the default case maps unrecognized roles to the literal string `"model"`, which is not a registered agent name.

## Goals / Non-Goals

**Goals:**
- Eliminate race condition in browser initialization with thread-safe lazy init
- Ensure failed browser connections are retried on next call
- Map ADK's `"model"` role to the correct agent name in event replay

**Non-Goals:**
- Refactoring the entire browser tool architecture
- Persisting Author field in session store (separate concern)

## Decisions

### Decision 1: sync.Once with retry-on-failure for browser init

**Choice**: Use `sync.Once` to serialize `initBrowser()`. On failure, reset the `Once` so the next call retries. Only assign `t.browser` after successful `Connect()`.

**Rationale**: `sync.Once` is the standard Go pattern for thread-safe lazy initialization. The reset-on-failure pattern ensures transient errors (Chrome not yet ready, network issues) are retried instead of permanently cached. Assigning `t.browser` only after `Connect()` succeeds prevents the partial-initialization bug.

**Alternative considered**: Full mutex around `ensureBrowser()` — works but heavier than needed since initialization is a one-time operation.

### Decision 2: Add "model" to assistant case in role-to-author switch

**Choice**: Add `"model"` alongside `"assistant"` in the switch case, and change the `default` fallback to use `rootAgentName` instead of literal `"model"`.

**Rationale**: The ADK framework uses `role: "model"` for LLM-generated responses. This is semantically equivalent to `"assistant"` — both represent model output. The default fallback should also map to a valid agent name to prevent any future unknown-role issues.

## Risks / Trade-offs

- **sync.Once reset**: Reassigning `sync.Once{}` after failure is safe but unconventional — well-documented in code comments to prevent confusion.
- **Close() must reset Once**: The `Close()` method must reset `browserOnce` to allow re-initialization after browser cleanup.
