## 1. Config & Event Types

- [x] 1.1 Add `ConfirmationDepth uint64` field to `EscrowOnChainConfig` in `internal/config/types_economy.go`
- [x] 1.2 Add `EscrowReorgDetectedEvent` struct to `internal/eventbus/economy_events.go`

## 2. EventMonitor Core Logic

- [x] 2.1 Extract `BlockchainClient` interface from `*ethclient.Client` in `internal/economy/escrow/hub/monitor.go`
- [x] 2.2 Add `confirmationDepth`, `blockHashes`, `maxHashCache` fields to EventMonitor struct
- [x] 2.3 Add `WithConfirmationDepth` MonitorOption
- [x] 2.4 Modify `fetchAndPublish()` to calculate safeBlock and apply confirmation depth buffer
- [x] 2.5 Add reorg detection logic (safeBlock < lastBlock → rollback + event publish)
- [x] 2.6 Add block hash caching and continuity check for silent reorg detection
- [x] 2.7 Add `trimBlockHashCache()` for bounded cache size

## 3. Wiring & Bridge

- [x] 3.1 Wire `ConfirmationDepth` config → `WithConfirmationDepth` option in `internal/app/wiring_economy.go`
- [x] 3.2 Subscribe to `EscrowReorgDetectedEvent` in `internal/app/bridge_onchain_escrow.go` with ERROR/WARN logging

## 4. CLI & TUI

- [x] 4.1 Display `ConfirmationDepth` in `lango economy escrow show` output in `internal/cli/economy/escrow.go`
- [x] 4.2 Add `ConfirmationDepth` input field to on-chain escrow TUI form in `internal/cli/settings/forms_economy.go`

## 5. Tests

- [x] 5.1 Add `mockBlockchainClient` to test file
- [x] 5.2 Test: `TestConfirmationDepth_ToBlockCalculation` — depth=2, latest=100 → safeBlock=98
- [x] 5.3 Test: `TestConfirmationDepth_Zero` — depth=0 → processes up to latest
- [x] 5.4 Test: `TestReorgDetection_Rollback` — shallow reorg within depth, ExceedsDepth=false
- [x] 5.5 Test: `TestReorgDetection_DeepReorg` — deep reorg exceeding depth, ExceedsDepth=true
- [x] 5.6 Test: `TestBlockHashCache_Trim` — cache trimming when exceeding maxHashCache
- [x] 5.7 Test: `TestWithConfirmationDepth_Option` — option sets confirmationDepth correctly

## 6. Verification

- [x] 6.1 `go build ./...` passes
- [x] 6.2 `go test ./internal/economy/escrow/hub/...` passes
- [x] 6.3 `go test ./...` passes (no regressions)
