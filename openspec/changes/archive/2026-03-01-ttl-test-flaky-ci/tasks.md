## 1. Fix TTL Test Timing

- [x] 1.1 Update `TestEntStore_TTL` TTL from `1ms` to `50ms` and sleep from `5ms` to `100ms`
- [x] 1.2 Update `TestEntStore_TTL_DeleteAndRecreate` TTL from `1ms` to `50ms` and sleep from `5ms` to `100ms`

## 2. Verification

- [x] 2.1 Run `CGO_ENABLED=1 go test -race -count=10 ./internal/session/` to confirm no flakiness
