## Why

`lango serve` can report a successful startup even when the gateway cannot bind its configured address. The application lifecycle currently starts the gateway in a background goroutine and returns success before the gateway's `ListenAndServe` error is available, so immediate bind failures are reduced to logs instead of becoming startup failures.

## What Changes

- Bind the gateway listener before the application lifecycle reports the gateway component as started.
- Preserve the gateway server's existing blocking `Start()` contract and graceful shutdown behavior.
- Add regression coverage for occupied-port startup failures and serve output suppression on startup errors.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `gateway-server`: Application startup propagates immediate gateway listen failures and `lango serve` does not print a startup summary after startup fails.

## Impact

- Affected code: `internal/gateway/server.go`, `internal/app/app.go`, `cmd/lango/main.go` tests.
- Affected tests: gateway lifecycle tests, application lifecycle tests, serve command output tests.
- Affected specs: `openspec/specs/gateway-server/spec.md`.
