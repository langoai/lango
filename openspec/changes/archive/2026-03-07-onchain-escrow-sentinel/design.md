## Context

The Lango P2P agent economy uses `internal/economy/escrow/` with a custodian model (USDCSettler). The existing escrow engine has a complete state machine (pending→funded→active→completed→released + disputed/expired/refunded) and `SettlementExecutor` interface (Lock/Release/Refund). The `contract.Caller` provides gas estimation, nonce management, and retry logic for on-chain interactions. The `eventbus.Bus` handles typed synchronous event distribution.

## Goals / Non-Goals

**Goals:**
- Trustless on-chain escrow via smart contracts on Base network
- Dual-mode settlement: Hub (multi-deal, gas-efficient) and Vault (per-deal isolation, EIP-1167 clones)
- Reuse existing `contract.Caller` and `SettlementExecutor` interface
- Security monitoring with anomaly detection
- Full backward compatibility with existing custodian mode

**Non-Goals:**
- Cross-chain escrow (Base only)
- Automated arbitration (human arbitrator resolves disputes)
- Real-time WebSocket event streaming (polling-based only)
- Token swaps or DEX integration

## Decisions

### AD-1: Dual-mode settlement (Hub vs Vault)
Hub mode stores all deals in a single contract (gas-efficient for high-volume). Vault mode creates per-deal EIP-1167 minimal proxy clones (deal isolation, composability). Config selects mode via `economy.escrow.onChain.mode`. Both implement `SettlementExecutor`.

**Alternative**: Single contract only. Rejected because per-deal isolation is important for high-value transactions.

### AD-2: Typed clients wrapping contract.Caller
HubClient, VaultClient, FactoryClient wrap `contract.Caller` for type-safe operations. Reuses all gas estimation, nonce management, and retry logic.

**Alternative**: Direct go-ethereum bindings (abigen). Rejected to avoid code generation dependency and maintain consistency with existing `contract.Caller` patterns.

### AD-3: ABI embedding via go:embed
ABI JSON files committed to `internal/economy/escrow/hub/abi/` and embedded at compile time. No runtime file loading.

### AD-4: Ent schema for persistent escrow tracking
`escrow_deals` table with on-chain mapping fields (chain_id, hub_address, on_chain_deal_id, tx hashes). EntStore implements existing `escrow.Store` interface with additional on-chain methods.

### AD-5: Polling-based event monitor
`eth_getLogs` polling with configurable interval (default 15s). Publishes typed events to `eventbus.Bus`. Simpler than WebSocket subscriptions and works with all RPC providers.

### AD-6: Sentinel engine with pluggable detectors
Detector interface allows adding new anomaly patterns. Engine subscribes to eventbus events and stores alerts in memory. 5 initial detectors cover common attack patterns.

### AD-7: Additive config under economy.escrow.onChain
All new configuration is under `economy.escrow.onChain` sub-struct. Existing custodian mode config unchanged. `onChain.enabled=false` is the default.

## Risks / Trade-offs

- [Polling latency] Event monitor has up to `pollInterval` delay. → Acceptable for escrow operations (not time-critical).
- [In-memory sentinel alerts] Alerts lost on restart. → Acceptable for MVP; can add Ent persistence later.
- [No automated dispute resolution] Disputes require manual arbitrator intervention. → By design for trust and legal compliance.
- [EIP-1167 clone gas overhead] Each vault creation costs ~45k gas for proxy deployment. → Acceptable; per-deal isolation justifies cost.
