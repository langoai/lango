## Why

Code review of the adaptive-idle-timeout implementation revealed two issues: (1) `initGateway()` in `wiring.go` duplicated timeout resolution logic with incomplete handling (missing `AutoExtendTimeout` legacy, `IdleTimeout < 0`, and ceiling≤idle 3x fallback cases), and (2) `gateway/server.go` used raw string literals for error types instead of the `deadline.Reason` constants.

## What Changes

- Extract `resolveTimeouts()` logic from `internal/app/channels.go` into a reusable `deadline.ResolveTimeouts()` package-level function with `TimeoutConfig` input struct
- Replace inline timeout computation in `initGateway()` (`wiring.go`) with `deadline.ResolveTimeouts()` call — fixing the incomplete logic
- Replace raw string error types (`"timeout"`, `"idle_timeout"`, `"max_timeout"`) in `gateway/server.go` with `string(deadline.ReasonXxx)` constants
- Move 8 `TestResolveTimeouts_*` tests from `internal/app/channels_test.go` to `internal/deadline/resolve_test.go`; simplify `channels_test.go` to a single delegation integration test

## Capabilities

### New Capabilities

(none — this is a refactor of existing functionality)

### Modified Capabilities

- `adaptive-idle-timeout`: Timeout resolution logic is now a shared package-level function (`deadline.ResolveTimeouts`) rather than an app method, and gateway error types use typed constants instead of raw strings.

## Impact

- `internal/deadline/resolve.go` (new file) — `ResolveTimeouts()` + `TimeoutConfig`
- `internal/deadline/resolve_test.go` (new file) — 8 tests moved from app package
- `internal/app/channels.go` — `resolveTimeouts()` now delegates to `deadline.ResolveTimeouts()`
- `internal/app/channels_test.go` — simplified to 1 integration test
- `internal/app/wiring.go` — `initGateway()` uses `deadline.ResolveTimeouts()` instead of inline logic
- `internal/gateway/server.go` — error type strings replaced with `deadline.Reason` constants
- No API changes, no breaking changes, no dependency additions
