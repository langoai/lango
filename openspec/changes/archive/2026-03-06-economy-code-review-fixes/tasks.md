## 1. Code Reuse — parseUSDC Deduplication

- [x] 1.1 Remove private `parseUSDC` from `internal/economy/budget/engine.go`, import and use `wallet.ParseUSDC` with sign check
- [x] 1.2 Remove private `parseUSDC` from `internal/economy/risk/engine.go`, import and use `wallet.ParseUSDC` with fallback to default
- [x] 1.3 Update `internal/economy/risk/engine_test.go` TestParseUSDC to use `wallet.ParseUSDC` signature

## 2. Quality — Stringly-typed Action Constants

- [x] 2.1 Replace raw string action comparisons in `internal/app/wiring_economy.go` with `negotiation.ActionPropose/Counter/Accept/Reject` constants

## 3. Quality — Compile-time Interface Check

- [x] 3.1 Add `var _ escrow.SettlementExecutor = (*noopSettler)(nil)` in `internal/app/wiring_economy.go`

## 4. Efficiency — Lock-held Callback Fix

- [x] 4.1 Refactor `checkThresholds` in `internal/economy/budget/engine.go` to collect triggered thresholds under lock, fire callbacks after unlock

## 5. Efficiency — Capacity Hints

- [x] 5.1 Add capacity hint `make([]*agent.Tool, 0, 12)` in `internal/app/tools_economy.go`
- [x] 5.2 Add capacity hint to `ListByPeer` result slice in `internal/economy/escrow/store.go`

## 6. Verification

- [x] 6.1 Run `go build ./...` — build passes
- [x] 6.2 Run `go test ./internal/economy/...` — all tests pass
- [x] 6.3 Run `go test ./internal/app/...` — all tests pass
