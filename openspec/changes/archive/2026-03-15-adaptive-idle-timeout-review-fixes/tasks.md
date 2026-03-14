## 1. Extract ResolveTimeouts to deadline package

- [x] 1.1 Create `internal/deadline/resolve.go` with `TimeoutConfig` struct and `ResolveTimeouts()` function
- [x] 1.2 Create `internal/deadline/resolve_test.go` with 8 tests covering all resolution cases

## 2. Update channel handler delegation

- [x] 2.1 Replace `App.resolveTimeouts()` body in `internal/app/channels.go` with delegation to `deadline.ResolveTimeouts()`
- [x] 2.2 Simplify `internal/app/channels_test.go` to single integration test verifying delegation

## 3. Fix gateway timeout resolution

- [x] 3.1 Replace inline timeout computation in `initGateway()` (`internal/app/wiring.go`) with `deadline.ResolveTimeouts()` call
- [x] 3.2 Remove unused `time` import from `wiring.go`, add `deadline` import

## 4. Fix gateway error type strings

- [x] 4.1 Replace raw string error types in `internal/gateway/server.go` `handleChatMessage()` with `string(deadline.ReasonXxx)` constants

## 5. Verification

- [x] 5.1 `go build ./...` passes
- [x] 5.2 `go test ./internal/deadline/...` — all ResolveTimeouts tests pass
- [x] 5.3 `go test ./internal/app/...` — delegation integration test passes
- [x] 5.4 `go test ./internal/gateway/...` — all gateway tests pass
