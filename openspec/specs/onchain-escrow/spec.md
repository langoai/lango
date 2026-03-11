# On-Chain Escrow Sentinel Architecture

## Purpose

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
      confirmationDepth: 2     # blocks to wait for reorg protection (default: 2)
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
The system SHALL embed compiled ABI JSON files via `//go:embed` in `internal/economy/escrow/hub/abi/`. HubClient, VaultClient, FactoryClient, HubV2Client SHALL wrap `contract.Caller` for type-safe contract interaction. V2 ABI files (`LangoEscrowHubV2.abi.json`, `LangoVaultV2.abi.json`) SHALL be embedded alongside V1 ABIs.

#### Scenario: HubClient creates a deal
- **WHEN** HubClient.CreateDeal is called with seller, token, amount, and deadline
- **THEN** it calls contract.Caller.Write with the createDeal ABI method and returns the deal ID and tx hash

#### Scenario: FactoryClient creates a vault
- **WHEN** FactoryClient.CreateVault is called with seller, token, amount, deadline, and arbitrator
- **THEN** it calls the factory contract and returns VaultInfo with vault address and tx hash

#### Scenario: V2 ABI accessor functions
- **WHEN** `HubV2ABIJSON()` or `VaultV2ABIJSON()` is called
- **THEN** it SHALL return the raw ABI JSON string for the respective V2 contract

### Requirement: Dual-mode settlement executors
The system SHALL provide HubSettler and VaultSettler implementing the existing `SettlementExecutor` interface (Lock/Release/Refund). Config field `economy.escrow.onChain.mode` SHALL select between "hub" and "vault" modes. HubSettler SHALL additionally support V2 deal correlation via `SetDealMappingByDID(did, dealID)` for mapping DIDs to on-chain deal IDs.

#### Scenario: Hub mode settlement
- **WHEN** config has `economy.escrow.onChain.mode=hub` and `hubAddress` is set
- **THEN** selectSettler returns a HubSettler that uses HubClient for on-chain operations

#### Scenario: Vault mode settlement
- **WHEN** config has `economy.escrow.onChain.mode=vault` with factory and implementation addresses
- **THEN** selectSettler returns a VaultSettler that creates per-deal vault clones

#### Scenario: Fallback to custodian
- **WHEN** on-chain mode is enabled but required addresses are missing
- **THEN** selectSettler falls back to existing USDCSettler with a warning log

#### Scenario: V2 deal mapping by DID
- **WHEN** `HubSettler.SetDealMappingByDID(did, dealID)` is called
- **THEN** the settler SHALL store the DID-to-dealID mapping and `GetDealID(did)` SHALL return the stored dealID

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

### Requirement: On-chain escrow configuration with confirmation depth
The EscrowOnChainConfig SHALL include a `ConfirmationDepth` field (uint64) that specifies the number of blocks to wait before processing on-chain events. When the value is 0 or unset, the system SHALL use a default of 2 blocks for Base L2 reorg protection.

#### Scenario: Config with explicit confirmation depth
- **WHEN** `economy.escrow.onChain.confirmationDepth` is set to 5
- **THEN** the EventMonitor SHALL use confirmationDepth=5

#### Scenario: Config with zero confirmation depth
- **WHEN** `economy.escrow.onChain.confirmationDepth` is 0 or unset
- **THEN** the system SHALL apply a default of 2 blocks

### Requirement: Reorg event alerting in escrow bridge
The on-chain escrow bridge SHALL subscribe to `EscrowReorgDetectedEvent` and log the event. Deep reorgs (ExceedsDepth=true) SHALL be logged at ERROR level with CRITICAL prefix. Shallow reorgs SHALL be logged at WARN level.

#### Scenario: Deep reorg logging
- **WHEN** EscrowReorgDetectedEvent is received with ExceedsDepth=true
- **THEN** the bridge SHALL log at ERROR level including previousBlock, newBlock, and depth

#### Scenario: Shallow reorg logging
- **WHEN** EscrowReorgDetectedEvent is received with ExceedsDepth=false
- **THEN** the bridge SHALL log at WARN level including previousBlock, newBlock, and depth

### Requirement: Escrow agent tools
The system SHALL provide 10 escrow tools: escrow_create, escrow_fund, escrow_activate, escrow_submit_work, escrow_release, escrow_refund, escrow_dispute, escrow_resolve, escrow_status, escrow_list. State-changing tools SHALL be marked as dangerous.

#### Scenario: Agent creates and funds escrow
- **WHEN** agent calls escrow_create with seller DID and amount, then escrow_fund with the escrow ID
- **THEN** escrow is created in funded state with on-chain deposit if hub/vault mode is active

### Requirement: Expanded CLI commands
The system SHALL provide: `lango economy escrow list` (config summary), `lango economy escrow show` (detailed on-chain config), `lango economy escrow sentinel status` (sentinel health).

#### Scenario: CLI shows on-chain config
- **WHEN** user runs `lango economy escrow show`
- **THEN** system displays hub address, vault factory, arbitrator, token address, poll interval, and confirmation depth

### Requirement: On-chain escrow documentation in economy.md
The system SHALL include documentation for on-chain escrow (Hub/Vault dual-mode) in `docs/features/economy.md`, covering deal lifecycle, contract architecture, and configuration.

#### Scenario: Hub vs Vault mode documentation
- **WHEN** a user reads the on-chain escrow section in economy.md
- **THEN** they find descriptions of Hub mode (single contract, multiple deals) and Vault mode (per-deal EIP-1167 proxy)

#### Scenario: On-chain config keys in configuration.md
- **WHEN** a user reads `docs/configuration.md`
- **THEN** all 10 on-chain escrow config keys (`economy.escrow.onChain.*`, `economy.escrow.settlement.*`) are documented with types and defaults

### Requirement: On-chain escrow CLI documentation
The system SHALL document `escrow list`, `escrow show`, and `escrow sentinel status` CLI commands in `docs/cli/economy.md`.

#### Scenario: CLI command reference
- **WHEN** a user reads `docs/cli/economy.md`
- **THEN** they find usage, flags, and output examples for `lango economy escrow list`, `lango economy escrow show`, and `lango economy escrow sentinel status`

### Requirement: Escrow tools in system prompts
The system SHALL list all 10 `escrow_*` tools with correct names and workflow guidance in `prompts/TOOL_USAGE.md`.

#### Scenario: Tool names match code
- **WHEN** the agent reads TOOL_USAGE.md
- **THEN** tool names match those registered in `internal/app/tools_escrow.go`: `escrow_create`, `escrow_fund`, `escrow_activate`, `escrow_submit_work`, `escrow_release`, `escrow_refund`, `escrow_dispute`, `escrow_resolve`, `escrow_status`, `escrow_list`

### Requirement: Contracts documentation
The system SHALL document Foundry-based escrow contracts (LangoEscrowHub, LangoVault, LangoVaultFactory) in `docs/features/contracts.md`.

#### Scenario: Contract architecture documented
- **WHEN** a user reads `docs/features/contracts.md`
- **THEN** they find contract descriptions, deal states, events, and Foundry build/test commands

### Requirement: On-chain escrow events in economy.md
The system SHALL document the 6 new on-chain events in the Events Summary table of `docs/features/economy.md`.

#### Scenario: Events table updated
- **WHEN** a user reads the Events Summary in economy.md
- **THEN** events for DealCreated, DealDeposited, WorkSubmitted, DealReleased, DealRefunded, DealDisputed are listed

### Requirement: README reflects on-chain escrow
The system SHALL mention on-chain Hub/Vault escrow, Foundry contracts, and escrow CLI commands in `README.md`.

#### Scenario: Feature bullets updated
- **WHEN** a user reads README.md features section
- **THEN** on-chain escrow and Foundry contracts are mentioned

### Requirement: V2 event handling in EventMonitor
The EventMonitor SHALL detect and correctly parse V2 contract events that include refId as an extra indexed topic. V2 events use a 4-topic layout `[sig, refId, dealId, addr]` compared to V1's 3-topic layout `[sig, dealId, addr]`.

#### Scenario: V2 deposit event detection
- **WHEN** a V2 Deposited event is emitted with 4 topics `[sig, refId, dealId, buyer]`
- **THEN** the EventMonitor SHALL detect it as a V2 event via `isV2Event()` and extract dealID from topic index 2 and buyer from topic index 3

#### Scenario: V1 event backward compatibility
- **WHEN** a V1 Deposited event is emitted with 3 topics `[sig, dealId, buyer]`
- **THEN** the EventMonitor SHALL handle it as a V1 event and extract dealID from topic index 1 and buyer from topic index 2

#### Scenario: V2-only event names
- **WHEN** events named `SettlementFinalized`, `EscrowOpened`, or `MilestoneReached` are received
- **THEN** the EventMonitor SHALL treat them as V2 events unconditionally

#### Scenario: DisputeRaised vs Disputed event distinction
- **WHEN** a `DisputeRaised` event is received (V2 naming)
- **THEN** the EventMonitor SHALL treat it as a V2 dispute event with refId at topic index 1 and dealID at topic index 2

### Requirement: DanglingDetector periodic scan
The `DanglingDetector` SHALL periodically scan for escrows stuck in `Pending` status beyond `maxPending` duration and expire them. The scan SHALL use `Store.ListByStatus(StatusPending)` instead of loading all escrows via `Store.List()`.

#### Scenario: Scan expires old pending escrows
- **WHEN** the scan runs and an escrow has been in `Pending` status longer than `maxPending`
- **THEN** the detector SHALL call `Engine.Expire` on that escrow and publish an `EscrowDanglingEvent`

#### Scenario: Scan skips non-pending escrows
- **WHEN** the scan runs
- **THEN** the detector SHALL NOT load or iterate escrows in non-pending statuses

#### Scenario: Configurable scan parameters
- **WHEN** `DanglingDetector` is created with `WithScanInterval(d)` and `WithMaxPending(d)` options
- **THEN** the detector SHALL scan at the specified interval and use the specified max pending threshold

#### Scenario: Lifecycle management
- **WHEN** `DanglingDetector.Start()` and `DanglingDetector.Stop()` are called
- **THEN** the detector SHALL start and stop its background scan goroutine gracefully

### Requirement: EscrowDanglingEvent for stuck escrow alerting
The system SHALL define an `EscrowDanglingEvent` type in `internal/eventbus/economy_events.go` published when a dangling escrow is detected and expired.

#### Scenario: Event fields
- **WHEN** an `EscrowDanglingEvent` is published
- **THEN** it SHALL contain EscrowID, BuyerDID, SellerDID, Amount, PendingSince, and Action fields

#### Scenario: Event name
- **WHEN** `EscrowDanglingEvent.EventName()` is called
- **THEN** it SHALL return `"escrow.dangling"`

### Requirement: Monitor V1/V2 topic offset helpers
The `EventMonitor` SHALL use helper methods to extract deal ID and address from log topics, abstracting the V1/V2 topic offset difference.

#### Scenario: extractDealAndAddress for V1 events
- **WHEN** a V1 event log is processed (3 topics: [sig, dealId, addr])
- **THEN** `extractDealAndAddress` SHALL return `topicToBigInt(log, 1)` as dealID and `topicToAddress(log, 2)` as address

#### Scenario: extractDealAndAddress for V2 events
- **WHEN** a V2 event log is processed (4 topics: [sig, refId, dealId, addr])
- **THEN** `extractDealAndAddress` SHALL return `topicToBigInt(log, 2)` as dealID and `topicToAddress(log, 3)` as address

#### Scenario: extractDealID for resolution events
- **WHEN** a DealResolved or SettlementFinalized event is processed
- **THEN** `extractDealID` SHALL return the correct dealID regardless of V1/V2 layout

