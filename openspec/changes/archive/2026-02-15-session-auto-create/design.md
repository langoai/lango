## Context

The ADK Runner calls `SessionService.Get()` before processing any user input. When a channel (Telegram, Discord, Slack) or gateway receives the first message from a user, no session exists for that key yet. The current implementation returns an error, breaking the conversation flow. All channel handlers use the same `runAgent → RunAndCollect → Runner.Run` path, so the fix must apply at the adapter layer to benefit all callers.

## Goals / Non-Goals

**Goals:**
- First messages through any channel start a conversation without errors
- Existing sessions continue to work as before
- No code changes required in channel handlers or the gateway

**Non-Goals:**
- Populating auto-created sessions with channel metadata (ChannelType, ChannelID) — these are derived from the session key at runtime via `deriveChannelType()`
- Changing the underlying `EntStore.Get()` behavior — the store correctly returns "not found" errors
- Session lifecycle management (TTL, cleanup) changes

## Decisions

**1. Get-or-create at the ADK adapter layer, not the store layer**

The `SessionServiceAdapter.Get()` catches "session not found" errors from the store and delegates to `SessionServiceAdapter.Create()`. This keeps the store's behavior clean (Get returns errors for missing keys, which is correct for a data access layer) while adding the auto-creation logic at the application boundary where the ADK Runner interacts.

Alternative considered: Auto-create in `EntStore.Get()` — rejected because it would change the store contract for all callers, not just the ADK path.

Alternative considered: Pre-create sessions in each channel handler — rejected because it would require duplicating creation logic in every handler (Telegram, Discord, Slack, gateway).

**2. String-based error matching for "session not found"**

We use `strings.Contains(err.Error(), "session not found")` to detect the not-found case. This is pragmatic given the store returns `fmt.Errorf("session not found: %s", key)` without a typed error.

Alternative considered: Introduce a sentinel `ErrSessionNotFound` error — possible future improvement but out of scope for this fix.

## Risks / Trade-offs

- **[Risk] Accidental session creation on typos or invalid keys** → Low impact: sessions are lightweight and TTL-managed. Invalid sessions simply expire.
- **[Risk] String-based error matching is fragile** → Mitigated by the store's error message being stable internal code. A typed error would be more robust but is over-engineering for now.
- **[Trade-off] Auto-created sessions have no initial metadata** → Acceptable: channel metadata is derived from the session key pattern at runtime, not stored.
