## Why

Three parallel code review agents (Code Reuse, Quality, Efficiency) analyzed the P2P Economy Layer after its 45-task implementation was completed. This change addresses the high-value findings: duplicated utility functions, stringly-typed comparisons, missing compile-time checks, a lock-held callback risk, and missing capacity hints.

## What Changes

- Remove duplicate `parseUSDC()` from `budget/engine.go` and `risk/engine.go`; reuse `wallet.ParseUSDC()`
- Replace raw string comparisons (`"propose"`, `"counter"`, etc.) in `wiring_economy.go` with `negotiation.Action*` constants
- Add compile-time interface verification for `noopSettler` implementing `escrow.SettlementExecutor`
- Move `alertCallback` invocation outside mutex lock in `budget/engine.go:checkThresholds` to prevent potential blocking
- Add capacity hints to slice allocations in `tools_economy.go` and `escrow/store.go:ListByPeer`

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

(none — all changes are implementation-level refactoring with no spec-level behavior changes)

## Impact

- `internal/economy/budget/engine.go` — import `wallet`, replace private `parseUSDC`, refactor `checkThresholds` locking
- `internal/economy/risk/engine.go` — import `wallet`, replace private `parseUSDC`
- `internal/economy/risk/engine_test.go` — update `TestParseUSDC` to use `wallet.ParseUSDC` signature
- `internal/app/wiring_economy.go` — use `negotiation.Action*` constants, add `noopSettler` interface check
- `internal/app/tools_economy.go` — capacity hint for tools slice
- `internal/economy/escrow/store.go` — capacity hint for `ListByPeer` result slice
