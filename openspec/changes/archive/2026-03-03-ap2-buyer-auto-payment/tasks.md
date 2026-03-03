## 1. Core Implementation

- [x] 1.1 Add `authToMap()` helper to `internal/app/tools_p2p.go` that serializes `eip3009.Authorization` into paygate-compatible map format
- [x] 1.2 Add `buildP2PPaidInvokeTool()` to `internal/app/tools_p2p.go` with full buyer-side paid invocation flow
- [x] 1.3 Add required imports (`eip3009`, `contracts`, `common`, `hex`) to `tools_p2p.go`

## 2. Wiring

- [x] 2.1 Wire `buildP2PPaidInvokeTool(p2pc, pc)` in `internal/app/app.go` P2P tool registration block

## 3. Verification

- [x] 3.1 Verify `authToMap()` output format matches `paygate.parseAuthorization()` field expectations
- [x] 3.2 Run `go build ./...` — confirm compilation succeeds
- [x] 3.3 Run `go test ./internal/app/...` — confirm app tests pass
- [x] 3.4 Run `go test ./...` — confirm full regression tests pass
