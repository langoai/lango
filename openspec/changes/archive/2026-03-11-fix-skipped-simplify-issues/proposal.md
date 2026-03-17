## Why

A `/simplify` code review flagged 6 issues that were initially skipped as "out of scope" or "low impact." Re-evaluation shows 4 of these are worth fixing: a P0 goroutine lifecycle bug (fire-and-forget violating Go guidelines), a P1 full-store scan inefficiency, and two P2 code duplication issues. The remaining 2 issues are intentionally skipped (semantic clarity and smart contract abstraction risk).

## What Changes

- **Fix fire-and-forget goroutine** in `bridge_team_budget.go` — add `context.Context` parameter and `select` on `ctx.Done()` to prevent post-shutdown store access (P0 correctness bug).
- **Add `ListByStatus` to escrow `Store` interface** — enables `DanglingDetector` to query only pending escrows instead of loading entire store every 5 minutes (P1 efficiency).
- **Export `NoopSettler` from escrow package** — consolidates 3 duplicate noop settler definitions into `escrow.NoopSettler` (P2 deduplication).
- **Extract V1/V2 topic offset helpers in `monitor.go`** — `extractDealAndAddress` and `extractDealID` replace 6 identical `if isV2` branches (P2 deduplication).

## Capabilities

### New Capabilities

_(none)_

### Modified Capabilities

- `economy-escrow`: Add `ListByStatus(status EscrowStatus)` to `Store` interface; export `NoopSettler` type.
- `onchain-escrow`: Extract V1/V2 topic offset helpers in event monitor; `DanglingDetector.scan()` uses `ListByStatus` instead of `List()`.
- `p2p-team-payment`: `wireTeamBudgetBridge` accepts `context.Context` for goroutine lifecycle management.

## Impact

- `internal/economy/escrow/store.go` — interface change (new method `ListByStatus`)
- `internal/economy/escrow/ent_store.go` — new `EntStore.ListByStatus` implementation
- `internal/economy/escrow/noop_settler.go` — new file
- `internal/economy/escrow/hub/dangling_detector.go` — simplified scan
- `internal/economy/escrow/hub/monitor.go` — extracted helpers, simplified switch
- `internal/app/bridge_team_budget.go` — ctx parameter, select pattern
- `internal/app/app.go` — app-level ctx/cancel for shutdown signalling
- `internal/app/types.go` — ctx/cancel fields on App
- `internal/app/wiring_economy.go` — removed local noopSettler
- Test files updated: `bridge_integration_test.go`, `bridge_onchain_escrow_test.go`, `dangling_detector_test.go`
