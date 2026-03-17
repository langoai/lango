## ADDED Requirements

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

### Requirement: Dangling escrow detector
The system SHALL provide a `DanglingDetector` component that periodically scans for escrows stuck in Pending status beyond a configurable threshold and automatically expires them.

#### Scenario: Stuck escrow detection
- **WHEN** an escrow has been in Pending status for longer than `maxPending` (default: 10 minutes)
- **THEN** the DanglingDetector SHALL call `engine.Expire()` on the escrow and publish an `EscrowDanglingEvent` to the event bus

#### Scenario: Healthy escrows unaffected
- **WHEN** a Pending escrow has been pending for less than `maxPending`
- **THEN** the DanglingDetector SHALL leave it untouched

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

## MODIFIED Requirements

### Requirement: Dual-mode settlement executors (MODIFIED)
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

### Requirement: Go ABI embedding and typed clients (MODIFIED)
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
