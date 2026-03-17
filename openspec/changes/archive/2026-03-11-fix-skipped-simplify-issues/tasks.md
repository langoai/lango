## 1. Fire-and-Forget Goroutine Fix (P0)

- [x] 1.1 Add `ctx context.Context` parameter to `wireTeamBudgetBridge` in `internal/app/bridge_team_budget.go`
- [x] 1.2 Replace bare goroutine with `select { case <-timer.C: releaseFn() case <-ctx.Done(): releaseFn() }` pattern
- [x] 1.3 Add `ctx`/`cancel` fields to `App` struct in `internal/app/types.go`
- [x] 1.4 Create context in `New()` and cancel in `Stop()` in `internal/app/app.go`
- [x] 1.5 Update call site in `app.go` to pass `app.ctx`
- [x] 1.6 Update 4 test call sites in `bridge_integration_test.go` to pass `context.Background()`

## 2. DanglingDetector ListByStatus (P1)

- [x] 2.1 Add `ListByStatus(status EscrowStatus) []*EscrowEntry` to `Store` interface in `internal/economy/escrow/store.go`
- [x] 2.2 Implement `ListByStatus` on `memoryStore` in `internal/economy/escrow/store.go`
- [x] 2.3 Implement `ListByStatus` on `EntStore` in `internal/economy/escrow/ent_store.go`
- [x] 2.4 Replace `dd.store.List()` with `dd.store.ListByStatus(escrow.StatusPending)` in `dangling_detector.go` and remove in-memory status filter

## 3. NoopSettler Export (P2)

- [x] 3.1 Create `internal/economy/escrow/noop_settler.go` with exported `NoopSettler` type and compile-time check
- [x] 3.2 Remove `noopSettler` type from `internal/app/wiring_economy.go`, use `escrow.NoopSettler{}`
- [x] 3.3 Remove `noopTestSettler` from `internal/economy/escrow/hub/dangling_detector_test.go`, use `escrow.NoopSettler{}`
- [x] 3.4 Update `bridge_onchain_escrow_test.go` to use `escrow.NoopSettler{}`
- [x] 3.5 Update `bridge_integration_test.go` to use `escrow.NoopSettler{}`

## 4. Monitor V1/V2 Topic Offset Helpers (P2)

- [x] 4.1 Add `extractDealAndAddress(log, isV2) (dealID, addr string)` helper to `monitor.go`
- [x] 4.2 Add `extractDealID(log, isV2) string` helper to `monitor.go`
- [x] 4.3 Simplify Deposited, WorkSubmitted, Released, Refunded cases to use `extractDealAndAddress`
- [x] 4.4 Simplify DealResolved, SettlementFinalized cases to use `extractDealID`

## 5. Verification

- [x] 5.1 `go build ./...` passes
- [x] 5.2 `go test ./internal/app/...` passes
- [x] 5.3 `go test ./internal/economy/escrow/...` passes
- [x] 5.4 `go test ./internal/economy/escrow/hub/...` passes
