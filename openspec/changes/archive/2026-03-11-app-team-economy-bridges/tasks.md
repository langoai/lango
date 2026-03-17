# Tasks: App Bridge Layer — Team-Economy Integration

## 1. Event Type Definitions

- [x] 1.1 Add on-chain escrow event types to `internal/eventbus/economy_events.go`: `EscrowOnChainDepositEvent`, `EscrowOnChainReleaseEvent`, `EscrowOnChainRefundEvent`, `EscrowOnChainDisputeEvent`, `EscrowOnChainResolvedEvent`, `EscrowDanglingEvent`
- [x] 1.2 Add team lifecycle event types to `internal/eventbus/team_events.go`: `TeamMemberUnhealthyEvent`, `TeamBudgetWarningEvent`, `TeamGracefulShutdownEvent`

## 2. On-Chain Escrow Bridge

- [x] 2.1 Implement `tryEscrowTransition` helper for idempotent state transitions with `ErrInvalidTransition` handling in `internal/app/bridge_onchain_escrow.go`
- [x] 2.2 Implement `initOnChainEscrowBridge` with subscribers for all five on-chain event types (deposit, release, refund, dispute, resolved)
- [x] 2.3 Implement `isAlreadyTransitioned` helper for error classification
- [x] 2.4 Add tests for each event type, idempotency, and empty escrow ID handling in `internal/app/bridge_onchain_escrow_test.go`

## 3. Team Reputation Bridge

- [x] 3.1 Implement `initTeamReputationBridge` with subscribers for `TeamMemberUnhealthyEvent`, `TeamTaskCompletedEvent`, and `ReputationChangedEvent` in `internal/app/bridge_team_reputation.go`
- [x] 3.2 Implement unhealthy member -> RecordTimeout -> score check -> KickMember logic
- [x] 3.3 Implement task completion -> RecordSuccess for active workers
- [x] 3.4 Implement reputation drop -> evict from all teams via `coordinator.TeamsForMember`
- [x] 3.5 Add tests for reputation adjustment and eviction logic in `internal/app/bridge_team_reputation_test.go`

## 4. Team Shutdown Bridge

- [x] 4.1 Implement `initTeamShutdownBridge` with subscribers for `BudgetAlertEvent` and `BudgetExhaustedEvent` in `internal/app/bridge_team_shutdown.go`
- [x] 4.2 Implement 80% threshold -> `TeamBudgetWarningEvent` publishing
- [x] 4.3 Implement budget exhaustion -> `coordinator.GracefulShutdown` trigger
- [x] 4.4 Add tests for warning threshold, below-threshold ignore, and shutdown paths in `internal/app/bridge_team_shutdown_test.go`

## 5. Economy Wiring Enhancements

- [x] 5.1 Update `selectSettler` in `internal/app/wiring_economy.go` to support `"hub"` mode with `HubSettler`
- [x] 5.2 Update `selectSettler` to support `"vault"` mode with `VaultSettler`
- [x] 5.3 Wire `DanglingDetector` creation when on-chain mode is enabled
- [x] 5.4 Wire `initOnChainEscrowBridge` call during economy initialization
- [x] 5.5 Register `EventMonitor` and `DanglingDetector` with lifecycle registry via `registerEconomyLifecycle`

## 6. P2P Wiring Enhancements

- [x] 6.1 Create `HealthMonitor` in `initP2P` with configurable interval and max missed heartbeats
- [x] 6.2 Add `healthMonitor` field to `p2pComponents` struct
- [x] 6.3 Wire bridge initialization calls from `app.go` (team-escrow and team-budget bridges)

## 7. Configuration Support

- [x] 7.1 Add on-chain escrow config fields (`Mode`, `VaultFactoryAddress`, `VaultImplementation`, `ArbitratorAddress`) to `internal/config/types_economy.go`
- [x] 7.2 Add team health config fields (`HealthCheckInterval`, `MaxMissedHeartbeats`) to `internal/config/types_p2p.go`
