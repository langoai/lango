# Design: App Bridge Layer — Team-Economy Integration

## Context

The Lango system has two independently evolving subsystems: the P2P team layer (`internal/p2p/team/`) and the economy layer (`internal/economy/`). Teams coordinate distributed agent work (formation, task delegation, health monitoring, disbandment), while the economy layer manages financial primitives (escrow, budget, pricing, risk). These subsystems communicate through the centralized `eventbus.Bus`.

Before this change, the subsystems were wired in isolation:
- On-chain escrow contract events (deposit, release, dispute) were monitored by `hub.EventMonitor` and published to the bus, but nothing reacted to them.
- Team task outcomes did not affect peer reputation.
- Team disbandment or budget exhaustion did not trigger escrow settlement.
- The escrow settler only supported custodian mode (USDCSettler).

## Goals / Non-Goals

**Goals:**
1. Wire on-chain escrow events to the local escrow engine so the off-chain state machine stays synchronized with blockchain state.
2. Update peer reputation based on team task outcomes (success boosts, timeout penalties) and auto-kick low-reputation members.
3. Trigger graceful team shutdown when budget is exhausted, including budget warning events at the 80% threshold.
4. Support hub and vault on-chain settlement modes alongside the existing custodian mode.
5. Wire DanglingDetector to expire stuck pending escrows.

**Non-Goals:**
1. Modifying the escrow engine state machine itself — bridges only call existing transition methods.
2. Adding new CLI commands or TUI surfaces for bridge management.
3. Changing the P2P protocol wire format.

## Decisions

### 1. Bridge Pattern: Event Subscription in `internal/app/`

All bridges live in `internal/app/bridge_*.go` as thin subscriber functions. Each bridge:
- Subscribes to specific event types via `eventbus.SubscribeTyped`.
- Calls existing engine/coordinator methods to perform side effects.
- Logs at debug level for idempotent/no-op cases, warn level for real errors.

**Rationale**: Bridges belong in the app layer because they cross subsystem boundaries (P2P <-> Economy). Placing them in `internal/app/` follows the existing pattern for `wireTeamEscrowBridge` and `wireTeamBudgetBridge`. The bridges contain no business logic — they are pure event-to-action translations.

### 2. Idempotent Transitions for On-Chain Bridge

The on-chain escrow bridge uses a `tryEscrowTransition` helper that catches `ErrInvalidTransition` and logs at debug level instead of warning. This makes the bridge safe for duplicate event delivery (e.g., EventMonitor replaying events after restart).

**Rationale**: On-chain events can be replayed during block reprocessing. Without idempotency, duplicate events would flood error logs. The helper centralizes this pattern for all five event types.

### 3. Reputation-Driven Member Eviction

The team reputation bridge subscribes to three event types:
- `TeamMemberUnhealthyEvent` -> `RecordTimeout` -> check score -> `KickMember` if below threshold.
- `TeamTaskCompletedEvent` -> `RecordSuccess` for active workers.
- `ReputationChangedEvent` -> check score -> `KickMember` from all teams if below threshold.

**Rationale**: A reactive approach (event-driven eviction) is simpler and more predictable than periodic polling. The configurable `minScore` threshold allows operators to tune sensitivity.

### 4. Budget-Driven Graceful Shutdown

The team shutdown bridge handles two thresholds:
- At 80% budget consumption: publishes `TeamBudgetWarningEvent` for UI/alerting.
- At 100% budget consumption: calls `coordinator.GracefulShutdown` which creates git bundles, settles payments, and disbands the team.

**Rationale**: Two-stage alerting (warning then shutdown) gives operators time to react. The 80% threshold is hardcoded as a sensible default; the warning event enables downstream consumers (CLI, webhooks) to surface alerts.

### 5. Multi-Mode Settler Selection

`selectSettler` now dispatches on `config.Economy.Escrow.OnChain.Mode`:
- `"hub"` -> `hub.NewHubSettler` (shared escrow contract).
- `"vault"` -> `hub.NewVaultSettler` (per-deal beacon proxy vault).
- Default -> `escrow.NewUSDCSettler` (custodian mode).

**Rationale**: Hub mode is simpler and cheaper for high-volume deals. Vault mode provides per-deal isolation for high-value transactions. Both require on-chain configuration (contract addresses). Custodian mode remains the fallback for users without on-chain infrastructure.

## Risks / Trade-offs

- **Event Ordering**: EventBus is synchronous and single-threaded per subscriber. If a bridge handler is slow (e.g., on-chain call), it can delay other subscribers. Mitigation: all bridge handlers perform only local state mutations or coordinator calls, not on-chain transactions.
- **Reputation Cascading**: A single bad event can cascade through multiple bridges (unhealthy -> timeout -> kick -> disband -> shutdown). Mitigation: each bridge logs its actions; the minimum score threshold acts as a circuit breaker.
- **Duplicate Event Delivery**: The on-chain bridge is designed for idempotency, but the reputation bridge is not (duplicate `RecordSuccess` calls inflate scores). Mitigation: the EventMonitor deduplicates at the source level using block number tracking.
