## Why

Concurrent Telegram messages arriving at startup trigger simultaneous session auto-creation for the same session key, causing a `UNIQUE constraint failed: sessions.key` error. The check-then-create pattern in `SessionServiceAdapter.Get()` is not atomic, so multiple goroutines pass the "not found" check and all attempt `Create()`.

## What Changes

- Refactor `SessionServiceAdapter.Get()` auto-create logic into a dedicated `getOrCreate()` helper
- When `Create()` fails with a UNIQUE constraint error, retry `Get()` to fetch the session created by the winning goroutine
- Add concurrent auto-create test with `uniqueMockStore` that simulates UNIQUE constraint errors

## Capabilities

### New Capabilities

### Modified Capabilities
- `session-auto-create`: Add concurrent safety requirement — when multiple goroutines auto-create the same session simultaneously, at most one creates it and the rest retrieve the already-created session without error.

## Impact

- `internal/adk/session_service.go` — new `getOrCreate()` method, modified `Get()` method
- `internal/adk/state_test.go` — new `uniqueMockStore` type and `TestSessionServiceAdapter_GetAutoCreate_Concurrent` test
