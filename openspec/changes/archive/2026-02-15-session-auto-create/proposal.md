## Why

When a user sends the first message through a channel (Telegram, Discord, Slack) or gateway, the ADK Runner calls `SessionService.Get()` to retrieve the session. Since no session exists yet for that key, it returns a "session not found" error, causing the channel handler to fail and surface an error to the user. Sessions must be auto-created on first access so conversations can begin without prior setup.

## What Changes

- `SessionServiceAdapter.Get()` now implements a **get-or-create** pattern: if the session is not found in the store, it automatically creates a new empty session and returns it instead of erroring.
- The test `TestSessionServiceAdapter_GetNotFound` is replaced by `TestSessionServiceAdapter_GetAutoCreate` to verify the new auto-creation behavior.

## Capabilities

### New Capabilities

- `session-auto-create`: Automatic session creation on first access via `SessionServiceAdapter.Get()`, enabling seamless first-message handling across all channels.

### Modified Capabilities

- `ent-session-store`: The ADK adapter layer now auto-creates sessions on `Get()` miss, changing the contract from "error on not found" to "create on not found".

## Impact

- **Code**: `internal/adk/session_service.go` (Get method), `internal/adk/state_test.go` (test update)
- **APIs**: All channel handlers (Telegram, Discord, Slack) and gateway `chat.message` benefit without code changes — they already call `RunAndCollect` which goes through `SessionServiceAdapter.Get()`.
- **Dependencies**: No new dependencies. Added `strings` import to `session_service.go`.
- **Systems**: Docker deployments need rebuild to pick up the fix.
