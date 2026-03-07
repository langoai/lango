## Why

The Lango P2P agent economy currently uses a custodian model where the agent wallet holds USDC directly. This requires trust in the agent operator. To enable trustless peer-to-peer transactions, we need on-chain escrow contracts on Base network with dual-mode settlement (Hub for multi-deal efficiency, Vault for per-deal isolation via EIP-1167 clones), event monitoring, and security anomaly detection.

## What Changes

- Add Solidity contracts: LangoEscrowHub (master hub), LangoVault (per-deal vault), LangoVaultFactory (EIP-1167 clone factory)
- Add Go ABI package with embedded ABIs and typed clients (HubClient, VaultClient, FactoryClient) wrapping existing `contract.Caller`
- Add HubSettler and VaultSettler as new `SettlementExecutor` implementations alongside existing USDCSettler
- Add Ent schema for persistent escrow deal tracking (replaces in-memory store for on-chain deals)
- Add polling-based EventMonitor that watches contract events via `eth_getLogs` and publishes to eventbus
- Add Security Sentinel engine with 5 anomaly detectors (rapid creation, large withdrawal, repeated disputes, unusual timing, balance drop)
- Add 10 escrow agent tools + 4 sentinel agent tools
- Add expanded CLI commands for escrow management and sentinel monitoring
- Add config under `economy.escrow.onChain` (fully additive, backward compatible)

## Capabilities

### New Capabilities
- `onchain-escrow`: On-chain escrow system with Hub and Vault dual-mode settlement, typed Go clients, settlement executors, event monitoring, and persistent Ent-backed storage
- `escrow-sentinel`: Security anomaly detection engine with 5 detectors, alert management, agent tools, and CLI monitoring commands

### Modified Capabilities
- `payment-service`: Added EscrowOnChainConfig sub-struct to EscrowConfig for on-chain settlement parameters
- `event-bus`: Added 6 on-chain escrow event types (deposit, work, release, refund, dispute, resolved)

## Impact

- **Config**: New `economy.escrow.onChain` section (additive, existing custodian mode unchanged)
- **Dependencies**: Uses existing `github.com/ethereum/go-ethereum` for ABI parsing and contract interaction
- **Database**: New `escrow_deals` Ent schema for persistent tracking
- **App wiring**: `selectSettler()` function in `wiring_economy.go`, sentinel engine lifecycle
- **CLI**: Expanded `lango economy escrow` with list/show/sentinel subcommands
- **Agent tools**: 14 new tools registered under "escrow" and "sentinel" catalog categories
