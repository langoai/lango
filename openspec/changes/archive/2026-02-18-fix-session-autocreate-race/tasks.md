## 1. Core Fix

- [x] 1.1 Extract `getOrCreate()` helper from `SessionServiceAdapter.Get()` in `internal/adk/session_service.go`
- [x] 1.2 In `getOrCreate()`, catch UNIQUE constraint errors from `Create()` and retry `store.Get()` to fetch the already-created session
- [x] 1.3 Update `Get()` to delegate both auto-create paths (err-based and nil-based) to `getOrCreate()`

## 2. Tests

- [x] 2.1 Add `uniqueMockStore` type in `internal/adk/state_test.go` that simulates UNIQUE constraint errors on duplicate Create
- [x] 2.2 Add `TestSessionServiceAdapter_GetAutoCreate_Concurrent` test with 10 goroutines racing to auto-create the same session
- [x] 2.3 Verify existing `TestSessionServiceAdapter_GetAutoCreate` still passes (single-goroutine path)

## 3. Verification

- [x] 3.1 Run `go build ./...` — no compilation errors
- [x] 3.2 Run `go test ./internal/adk/...` — all tests pass including new concurrent test
