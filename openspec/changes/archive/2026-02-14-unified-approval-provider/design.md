## Context

Tool execution approval currently flows through a single path: `wrapWithApproval` → `requestToolApproval` → `gateway.Server` (companion WebSocket) or TTY fallback. When a sensitive tool is invoked from a channel session (Telegram, Discord, Slack), the session key (e.g., `telegram:123:456`) carries the channel context but approval routing ignores it. If no companion is connected and stdin is not a TTY (which is always true in channel-originated requests), the request is unconditionally denied.

Each messaging platform provides native interactive components — Telegram InlineKeyboard, Discord Message Components, Slack Block Kit Actions — that can be used to prompt the originating user for approval within their native chat interface.

## Goals / Non-Goals

**Goals:**
- Introduce a `Provider` interface that abstracts approval request handling, decoupled from any specific transport
- Implement `CompositeProvider` that routes approval requests to the correct provider based on session key prefix
- Preserve exact existing behavior (gateway → TTY → deny) as the default path
- Enable each channel to register its own approval provider that uses native interactive components
- Maintain fail-closed semantics throughout the approval chain

**Non-Goals:**
- Multi-approver consensus (e.g., requiring N of M approvals) — single approver is sufficient
- Approval audit logging (can be added later as a decorator)
- Persistent approval state across restarts — approvals are ephemeral per-request
- Channel-to-channel approval delegation (e.g., Telegram user approving a Slack tool call)

## Decisions

### 1. Provider interface with session-key-based routing

**Decision**: Use a `Provider` interface with `CanHandle(sessionKey) bool` for routing, orchestrated by `CompositeProvider`.

**Rationale**: Session keys already encode channel origin (`telegram:chatID:userID`). Prefix matching gives O(n) routing with zero configuration. Alternative: explicit registry map — rejected because it couples provider registration to session key format knowledge in two places.

### 2. GatewayApprover interface instead of importing gateway package

**Decision**: Define a `GatewayApprover` interface in the `approval` package that `gateway.Server` already satisfies.

**Rationale**: Avoids circular dependency (`approval` → `gateway` → `approval`). The interface has only two methods (`HasCompanions`, `RequestApproval`), keeping the coupling surface minimal. Alternative: put everything in the `gateway` package — rejected because it would force channel packages to import gateway.

### 3. TTY as special fallback, not a prefix-matched provider

**Decision**: `TTYProvider.CanHandle()` always returns false. It is set via `SetTTYFallback()` on the composite.

**Rationale**: TTY is a last-resort fallback that applies when no channel-specific provider matches. It should never "claim" a session key. This separation makes the routing logic clearer and prevents TTY from accidentally intercepting channel requests.

### 4. Channel providers use sync.Map for pending approvals

**Decision**: Each channel `ApprovalProvider` uses `sync.Map` to store `requestID → chan bool` for pending approval requests.

**Rationale**: Approval requests are short-lived and concurrent. `sync.Map` provides lock-free reads for the common path (callback lookup) and safe concurrent writes. Alternative: `sync.Mutex` + `map` — works but more verbose for this use case.

### 5. Approval timeout as a config field

**Decision**: Add `ApprovalTimeoutSec` to `InterceptorConfig` with a default of 30 seconds.

**Rationale**: Different deployments may need different timeout values. The per-provider timeout defaults to 30s but can be overridden at config level.

## Risks / Trade-offs

- **[Race condition on callback]** If a callback arrives after timeout, the pending channel is already deleted → callback is silently dropped. This is acceptable behavior (user sees "timeout" message).
- **[Interface expansion]** Adding methods to channel SDK interfaces (`BotAPI`, `Session`, `Client`) is a minor breaking change for test mocks → Mitigated by updating existing mock types in the same PR.
- **[Message edit failures]** Editing the approval message to remove buttons may fail (message deleted, permissions changed) → Non-critical; logged as warning, does not affect approval result.
- **[Single approver per request]** The first matching provider handles the request. If the provider is temporarily unavailable, the request fails rather than falling through → This is intentional fail-closed behavior.
