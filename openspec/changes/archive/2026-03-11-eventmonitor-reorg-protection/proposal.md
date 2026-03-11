## Why

EventMonitor (`internal/economy/escrow/hub/monitor.go`) processes on-chain events immediately at the latest block via `eth_getLogs`. On Base L2, block reorganizations can cause already-processed escrow state transitions (Fund, Release, Refund) to diverge from actual on-chain state. A confirmation depth buffer and reorg detection mechanism are needed to prevent this inconsistency.

## What Changes

- Add `ConfirmationDepth` config field to `EscrowOnChainConfig` for configurable block confirmation buffer
- Add `EscrowReorgDetectedEvent` to the event bus for reorg alerting
- Extract `BlockchainClient` interface from concrete `*ethclient.Client` for testability
- Modify `fetchAndPublish()` to only process up to `latest - confirmationDepth` (safe block)
- Add reorg detection: when `safeBlock < lastBlock`, roll back and publish alert event
- Add block hash caching with continuity checks for silent reorg detection
- Wire confirmation depth from config to monitor via `WithConfirmationDepth` option
- Subscribe to reorg events in the on-chain escrow bridge for logging (CRITICAL for deep reorgs)
- Display confirmation depth in CLI `escrow show` and TUI settings form

## Capabilities

### New Capabilities
- `eventmonitor-reorg-protection`: Confirmation depth buffer and chain reorganization detection for the on-chain escrow event monitor

### Modified Capabilities
- `onchain-escrow`: Add ConfirmationDepth configuration and reorg-safe event processing to the on-chain escrow monitor

## Impact

- `internal/economy/escrow/hub/monitor.go`: Core logic changes (interface extraction, confirmation depth, reorg detection)
- `internal/config/types_economy.go`: New `ConfirmationDepth` field on `EscrowOnChainConfig`
- `internal/eventbus/economy_events.go`: New `EscrowReorgDetectedEvent` type
- `internal/app/wiring_economy.go`: Config-to-monitor option wiring
- `internal/app/bridge_onchain_escrow.go`: Reorg event logging subscription
- `internal/cli/economy/escrow.go`: CLI display
- `internal/cli/settings/forms_economy.go`: TUI form field
- `internal/economy/escrow/hub/monitor_test.go`: 6 new test cases with mock blockchain client
