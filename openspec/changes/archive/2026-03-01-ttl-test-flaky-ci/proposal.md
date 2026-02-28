## Why

`TestEntStore_TTL_DeleteAndRecreate` fails on Ubuntu CI but passes on macOS. The test uses a 1ms TTL, which is too short for CI environments with `-race` detector overhead. The session expires between `Create()` and `Get()`, causing `ErrSessionExpired`.

## What Changes

- Increase TTL from `1ms` to `50ms` in `TestEntStore_TTL` and `TestEntStore_TTL_DeleteAndRecreate`
- Increase corresponding sleep durations from `5ms` to `100ms` (2x margin for reliable expiration)

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

(none — this is a test-only timing fix with no spec-level behavior changes)

## Impact

- `internal/session/store_test.go`: Two test functions updated with wider timing margins
- No production code changes
- CI reliability improved on Ubuntu runners with `-race` flag
