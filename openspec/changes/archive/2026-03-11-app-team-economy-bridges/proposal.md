# Proposal: App Bridge Layer — Team-Economy Integration

## Why

The P2P team lifecycle (formation, task execution, disbandment) and the economy layer (escrow, budget, reputation) were independently implemented but lacked event-driven integration. Without bridges, on-chain escrow events had no effect on the local escrow engine, team task outcomes did not update peer reputation, and team shutdown did not trigger escrow settlement. These bridges close the gap so the two subsystems react to each other's lifecycle events automatically.

## What Changes

- **On-Chain Escrow Bridge**: Subscribes to on-chain escrow events (deposit, release, refund, dispute, resolved) published by the EventMonitor and triggers corresponding escrow engine state transitions (Fund, Activate, Release, Refund, Dispute, Resolve). All transitions are idempotent.
- **Team Reputation Bridge**: Subscribes to TeamMemberUnhealthyEvent and TeamTaskCompletedEvent. Records timeouts and successes in the reputation store, kicks members whose score drops below a configurable threshold, and reacts to ReputationChangedEvent to evict low-reputation peers from all active teams.
- **Team Shutdown Bridge**: Subscribes to BudgetAlertEvent (>=80% threshold) to publish TeamBudgetWarningEvent, and BudgetExhaustedEvent to trigger GracefulShutdown on the team coordinator.
- **Economy Wiring Enhancements**: selectSettler now supports hub/vault on-chain modes in addition to the existing custodian mode. DanglingDetector is wired to expire stuck pending escrows. initOnChainEscrowBridge is called during economy initialization when on-chain mode is enabled.
- **New Event Types**: EscrowOnChainDepositEvent, EscrowOnChainReleaseEvent, EscrowOnChainRefundEvent, EscrowOnChainDisputeEvent, EscrowOnChainResolvedEvent, EscrowDanglingEvent, TeamMemberUnhealthyEvent, TeamBudgetWarningEvent, TeamGracefulShutdownEvent.

## Capabilities

### New Capabilities
- `app-team-economy-bridges`: Event-driven bridges connecting P2P team lifecycle events to escrow engine state transitions, reputation adjustments, and budget-triggered graceful shutdown.

### Modified Capabilities
- `economy-wiring`: Hub/vault settler selection, DanglingDetector wiring, on-chain escrow bridge initialization during economy setup.

## Impact

- **internal/app/**: Three new bridge files (bridge_onchain_escrow.go, bridge_team_reputation.go, bridge_team_shutdown.go) plus corresponding test files.
- **internal/app/wiring_economy.go**: selectSettler supports hub/vault modes; DanglingDetector and EventMonitor wired; initOnChainEscrowBridge called.
- **internal/app/wiring_p2p.go**: HealthMonitor creation and registration; team coordinator wiring enhancements.
- **internal/eventbus/**: New on-chain escrow events and team lifecycle events.
- **internal/economy/escrow/hub/**: DanglingDetector, EventMonitor, HubSettler, VaultSettler additions.
