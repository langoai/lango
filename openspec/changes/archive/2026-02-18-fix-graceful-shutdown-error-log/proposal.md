## Why

During Docker container `stop` (SIGTERM), `http.ErrServerClosed` is logged at Error level even though it is a normal shutdown signal returned by Go's `http.Server` after `Shutdown()` is called. This creates false alarm noise in production logs and makes genuine errors harder to spot.

## What Changes

- Filter `http.ErrServerClosed` in `gateway.Server.Start()` so it returns `nil` instead of propagating the expected shutdown error
- Downgrade shutdown-phase cleanup error logs from `Errorw` to `Warnw` in `app.Stop()` and `main.go` shutdown handler, since resource cleanup failures at process exit are non-actionable warnings

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `gateway-server`: `Start()` now treats `http.ErrServerClosed` as a normal return (nil), not an error
- `server`: Shutdown cleanup errors logged at Warn level instead of Error level

## Impact

- `internal/gateway/server.go` — `Start()` method filters `http.ErrServerClosed`
- `internal/app/app.go` — `Stop()` cleanup errors downgraded to Warn
- `cmd/lango/main.go` — shutdown handler error downgraded to Warn
- No API changes, no behavioral changes beyond log level adjustment
