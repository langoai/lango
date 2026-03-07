# On-Chain Escrow Sentinel Architecture

## Overview

Trustless on-chain escrow system for Lango P2P agent economy on Base network. Implements a dual-mode settlement architecture (Hub + Vault) with event monitoring and security anomaly detection.

## Architecture

### Settlement Modes

| Mode       | Description                                | Config Key                           |
|------------|--------------------------------------------|--------------------------------------|
| `custodian`| Agent wallet holds USDC directly (default) | `economy.escrow.onChain.enabled=false` |
| `hub`      | Master LangoEscrowHub contract             | `economy.escrow.onChain.mode=hub`     |
| `vault`    | Per-deal EIP-1167 vault clones             | `economy.escrow.onChain.mode=vault`   |

### Smart Contracts

- **LangoEscrowHub** (`contracts/src/LangoEscrowHub.sol`) — Multi-deal escrow hub with arbitrator-based dispute resolution
- **LangoVault** (`contracts/src/LangoVault.sol`) — Single-deal vault, initializable for EIP-1167 cloning
- **LangoVaultFactory** (`contracts/src/LangoVaultFactory.sol`) — Factory creating minimal proxy vaults

### Go Packages

| Package | Role |
|---------|------|
| `internal/economy/escrow/hub/` | ABI embedding, typed clients (HubClient, VaultClient, FactoryClient), settlers |
| `internal/economy/escrow/sentinel/` | Anomaly detection engine with 5 detectors |
| `internal/economy/escrow/hub/monitor.go` | Event polling from on-chain contracts |

### Agent Tools

**Escrow Tools** (10): `escrow_create`, `escrow_fund`, `escrow_activate`, `escrow_submit_work`, `escrow_release`, `escrow_refund`, `escrow_dispute`, `escrow_resolve`, `escrow_status`, `escrow_list`

**Sentinel Tools** (4): `sentinel_status`, `sentinel_alerts`, `sentinel_config`, `sentinel_acknowledge`

### CLI Commands

```
lango economy escrow status     # Config display
lango economy escrow list       # Config summary with on-chain mode
lango economy escrow show       # Detailed on-chain config
lango economy escrow sentinel status  # Sentinel health
```

## Configuration

```yaml
economy:
  escrow:
    enabled: true
    onChain:
      enabled: true
      mode: "hub"              # "hub" | "vault"
      hubAddress: "0x..."
      vaultFactoryAddress: "0x..."
      vaultImplementation: "0x..."
      arbitratorAddress: "0x..."
      tokenAddress: "0x..."    # USDC contract
      pollInterval: 15s
```

## Security Sentinel

5 anomaly detectors:
1. **RapidCreation** — >5 deals from same peer in 1 minute
2. **LargeWithdrawal** — Single release > threshold
3. **RepeatedDispute** — >3 disputes from same peer in 1 hour
4. **UnusualTiming** — Deal created and released within <1 minute (wash trading)
5. **BalanceDrop** — Contract balance drops >50% in single block

Alerts have severity levels: Critical, High, Medium, Low.

## Event Flow

```
Contract Event → EventMonitor (eth_getLogs polling)
                → eventbus.Bus
                → Sentinel Engine (detectors)
                → Alert storage
```

## Dependencies

- `github.com/ethereum/go-ethereum` — ABI parsing, contract interaction
- `internal/contract.Caller` — Gas estimation, nonce management, retry logic
- `internal/eventbus.Bus` — Event distribution

## Requirements

### Requirement: Hub package clients accept ContractCaller interface
HubClient, VaultClient, FactoryClient, HubSettler, and VaultSettler constructors SHALL accept `contract.ContractCaller` interface instead of `*contract.Caller`.

#### Scenario: Constructors accept interface
- **WHEN** `NewHubClient`, `NewVaultClient`, `NewFactoryClient`, `NewHubSettler`, or `NewVaultSettler` is called
- **THEN** the `caller` parameter type SHALL be `contract.ContractCaller`

#### Scenario: Existing callers unaffected
- **WHEN** existing code passes `*contract.Caller` to hub package constructors
- **THEN** it SHALL compile without changes because `*Caller` satisfies `ContractCaller`

### Requirement: Solidity contracts for on-chain escrow
The system SHALL provide three Solidity contracts: LangoEscrowHub (master multi-deal hub), LangoVault (single-deal vault for EIP-1167 cloning), and LangoVaultFactory (minimal proxy factory). Contracts SHALL implement deal lifecycle: create, deposit, submitWork, release, refund, dispute, resolveDispute.

#### Scenario: Hub deal lifecycle
- **WHEN** a buyer creates a deal on LangoEscrowHub with seller address, token, amount, and deadline
- **THEN** a new deal is stored with status Created, and DealCreated event is emitted

#### Scenario: Vault creation via factory
- **WHEN** LangoVaultFactory.createVault is called with buyer, seller, token, amount, deadline, and arbitrator
- **THEN** an EIP-1167 minimal proxy clone of LangoVault is created and VaultCreated event is emitted

### Requirement: Go ABI embedding and typed clients
The system SHALL embed compiled ABI JSON files via `//go:embed` in `internal/economy/escrow/hub/abi/`. HubClient, VaultClient, and FactoryClient SHALL wrap `contract.Caller` for type-safe contract interaction.

#### Scenario: HubClient creates a deal
- **WHEN** HubClient.CreateDeal is called with seller, token, amount, and deadline
- **THEN** it calls contract.Caller.Write with the createDeal ABI method and returns the deal ID and tx hash

#### Scenario: FactoryClient creates a vault
- **WHEN** FactoryClient.CreateVault is called with seller, token, amount, deadline, and arbitrator
- **THEN** it calls the factory contract and returns VaultInfo with vault address and tx hash

### Requirement: Dual-mode settlement executors
The system SHALL provide HubSettler and VaultSettler implementing the existing `SettlementExecutor` interface (Lock/Release/Refund). Config field `economy.escrow.onChain.mode` SHALL select between "hub" and "vault" modes.

#### Scenario: Hub mode settlement
- **WHEN** config has `economy.escrow.onChain.mode=hub` and `hubAddress` is set
- **THEN** selectSettler returns a HubSettler that uses HubClient for on-chain operations

#### Scenario: Vault mode settlement
- **WHEN** config has `economy.escrow.onChain.mode=vault` with factory and implementation addresses
- **THEN** selectSettler returns a VaultSettler that creates per-deal vault clones

#### Scenario: Fallback to custodian
- **WHEN** on-chain mode is enabled but required addresses are missing
- **THEN** selectSettler falls back to existing USDCSettler with a warning log

### Requirement: Persistent escrow storage via Ent
The system SHALL provide an EntStore implementing the existing `escrow.Store` interface with additional on-chain tracking methods: SetOnChainDealID, GetByOnChainDealID, SetTxHash.

#### Scenario: Store and retrieve on-chain deal mapping
- **WHEN** SetOnChainDealID is called with escrowID and dealID
- **THEN** GetByOnChainDealID with that dealID returns the corresponding escrowID

### Requirement: Polling-based event monitor
The system SHALL provide an EventMonitor that polls `eth_getLogs` at configurable intervals (default 15s), decodes contract events using embedded ABIs, and publishes typed events to eventbus.Bus.

#### Scenario: Monitor detects deposit event
- **WHEN** a Deposited event is emitted on the hub contract
- **THEN** EventMonitor publishes EscrowOnChainDepositEvent to eventbus with deal ID, buyer, amount, and tx hash

### Requirement: Escrow agent tools
The system SHALL provide 10 escrow tools: escrow_create, escrow_fund, escrow_activate, escrow_submit_work, escrow_release, escrow_refund, escrow_dispute, escrow_resolve, escrow_status, escrow_list. State-changing tools SHALL be marked as dangerous.

#### Scenario: Agent creates and funds escrow
- **WHEN** agent calls escrow_create with seller DID and amount, then escrow_fund with the escrow ID
- **THEN** escrow is created in funded state with on-chain deposit if hub/vault mode is active

### Requirement: Expanded CLI commands
The system SHALL provide: `lango economy escrow list` (config summary), `lango economy escrow show` (detailed on-chain config), `lango economy escrow sentinel status` (sentinel health).

#### Scenario: CLI shows on-chain config
- **WHEN** user runs `lango economy escrow show`
- **THEN** system displays hub address, vault factory, arbitrator, token address, and poll interval
