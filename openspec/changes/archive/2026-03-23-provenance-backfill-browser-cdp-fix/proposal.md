## Why

Two runtime errors observed in production:
1. Provenance child session registration fails with "session node not found" because parent sessions loaded via `Get()` are never registered in the provenance tree (only `Create()` fires the observer).
2. Browser `navigate` tool fails with CDP error -32000 "Inspected target navigated or closed" and the recovery middleware only handles panics, not CDP errors.

## What Changes

- Make provenance `rootSessionObserver` idempotent (check existence before registering) and call it from `Get()` for existing sessions (missing-only backfill)
- Add `GetNode()` method to `SessionTree` to support existence checks
- Add CDP target error recovery to `WithBrowserRecovery` middleware, scoped to `browser_navigate` only (no side-effect tool retry)

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `session-auto-create`: Get() now calls rootSessionObserver for existing sessions to ensure provenance tree backfill
- `tool-browser`: WithBrowserRecovery middleware now handles CDP target error for browser_navigate with session-reset retry

## Impact

- **Code**: `internal/adk/session_service.go`, `internal/provenance/session_tree.go`, `internal/app/wiring.go`, `internal/toolchain/mw_browser.go`
- **APIs**: No external API changes
- **Dependencies**: None
- **Systems**: No deployment changes
