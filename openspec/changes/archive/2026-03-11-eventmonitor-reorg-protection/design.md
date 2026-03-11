## Context

EventMonitor polls `eth_getLogs` to watch on-chain escrow contract events (Deposited, Released, Refunded, etc.) and publishes them to the event bus, triggering escrow state transitions. Currently it processes events at the latest block immediately, which is unsafe on L2 chains like Base where short reorgs (1-3 blocks) are common.

The escrow state transitions (Fund, Release, Refund) are irreversible in the local escrow engine, so processing a reverted event causes permanent state inconsistency.

## Goals / Non-Goals

**Goals:**
- Prevent processing of events from blocks that may be reorged (confirmation depth buffer)
- Detect when a reorg has occurred and alert operators (reorg detection + event bus alert)
- Make the RPC dependency testable via interface extraction
- Maintain backward compatibility (zero-config works with sensible default)

**Non-Goals:**
- Event rollback mechanism (escrow Release/Refund are on-chain settlements, rollback is impossible)
- Subscription-based event watching (eth_subscribe) — polling is simpler and sufficient
- Finality tracking via L1 (overkill for Base L2 where reorgs are typically 1-2 blocks)

## Decisions

### Two-Layer Defense Strategy
**Decision**: Use confirmation depth as primary defense + reorg detection as safety net.
**Rationale**: Confirmation depth alone handles 99.9% of L2 reorgs. Reorg detection catches edge cases where the chain reorganizes deeper than the configured depth, allowing operator alerting without requiring complex event rollback.
**Alternative**: Full event rollback with undo log — rejected because escrow settlements are on-chain and cannot be locally reversed.

### BlockchainClient Interface
**Decision**: Extract `BlockchainClient` interface (`HeaderByNumber`, `FilterLogs`) from `*ethclient.Client`.
**Rationale**: Enables unit testing of confirmation depth and reorg detection logic without a real Ethereum node. `*ethclient.Client` already satisfies the interface, so no caller changes needed.

### Default Confirmation Depth = 2
**Decision**: Default to 2 blocks when `ConfirmationDepth` is 0 or unset.
**Rationale**: Base L2 produces blocks every 2 seconds with reorgs typically 1-2 blocks. Depth of 2 provides safety with only 4 seconds of latency. Matches the PollInterval zero-means-default pattern.

### Block Hash Cache for Silent Reorg Detection
**Decision**: Cache block hashes and verify continuity on each poll cycle.
**Rationale**: A reorg that doesn't change the block number (same-height reorg) won't be caught by the `safeBlock < lastBlock` check. Hash continuity verification detects these silent reorgs. Cache is bounded at 256 entries with LRU-style trimming.

## Risks / Trade-offs

- [Event latency increases by `confirmationDepth * blockTime`] → Acceptable: 2 blocks × 2s = 4s on Base L2
- [Block hash cache uses memory] → Mitigated: bounded at 256 entries (~8KB), trimmed on overflow
- [Deep reorg exceeding confirmation depth] → Mitigated: EscrowReorgDetectedEvent with ExceedsDepth=true triggers CRITICAL log; operators can investigate manually
- [RPC rate increase from hash checks] → Mitigated: one extra HeaderByNumber per poll cycle (negligible vs existing FilterLogs call)
