## 1. V2 Solidity Contracts

- [x] 1.1 Create `contracts/src/interfaces/ILangoEconomy.sol` shared economy interface
- [x] 1.2 Create `contracts/src/interfaces/ISettler.sol` settler strategy interface
- [x] 1.3 Create `contracts/src/LangoEscrowHubV2.sol` UUPS-upgradeable hub with refId, deal types, modular settlers
- [x] 1.4 Create `contracts/src/LangoVaultV2.sol` per-deal vault with milestone releases
- [x] 1.5 Create `contracts/src/LangoBeaconVaultFactory.sol` beacon proxy vault factory
- [x] 1.6 Create `contracts/src/settlers/DirectSettler.sol` immediate transfer settler
- [x] 1.7 Create `contracts/src/settlers/MilestoneSettler.sol` phased milestone release settler

## 2. V2 Deployment and Upgrade Scripts

- [x] 2.1 Create `contracts/script/DeployV2.s.sol` for fresh V2 deployment (hub proxy, settlers, beacon, factory)
- [x] 2.2 Create `contracts/script/UpgradeV2.s.sol` for upgrading existing hub proxy to new implementation

## 3. V2 Foundry Tests

- [x] 3.1 Create `contracts/test/LangoEscrowHubV2.t.sol` comprehensive hub V2 tests (~700 lines)
- [x] 3.2 Create `contracts/test/LangoVaultV2.t.sol` comprehensive vault V2 tests (~394 lines)

## 4. Go V2 ABI and Types

- [x] 4.1 Embed `abi/LangoEscrowHubV2.abi.json` and `abi/LangoVaultV2.abi.json` via `//go:embed` in `abi.go`
- [x] 4.2 Add `ParseHubV2ABI()` and `ParseVaultV2ABI()` parsing helpers and accessor functions
- [x] 4.3 Add `OnChainDealType` enum (Simple=0, Milestone=1, Team=2) with `String()` method to `types.go`
- [x] 4.4 Add `OnChainDealV2` struct extending `OnChainDeal` with DealType, RefId, Settler fields to `types.go`

## 5. Go HubV2Client

- [x] 5.1 Create `internal/economy/escrow/hub/client_v2.go` with `HubV2Client` struct and `NewHubV2Client` constructor
- [x] 5.2 Implement `CreateSimpleEscrow`, `CreateMilestoneEscrow`, `CreateTeamEscrow` methods with refId
- [x] 5.3 Implement `DirectSettle` for immediate token transfer without escrow
- [x] 5.4 Implement `CompleteMilestone` and `ReleaseMilestone` for milestone management
- [x] 5.5 Implement `Deposit`, `Release`, `Refund`, `Dispute`, `ResolveDispute` standard operations
- [x] 5.6 Implement `GetDealV2` and `NextDealID` read methods with `parseDealV2Result` helper
- [x] 5.7 Create `internal/economy/escrow/hub/client_v2_test.go` with unit tests

## 6. HubSettler V2 Support

- [x] 6.1 Add `SetDealMappingByDID(did, dealID)` method to `HubSettler` in `hub_settler.go`
- [x] 6.2 Update `dealMap` comment to clarify it tracks both escrow IDs and DIDs
- [x] 6.3 Update `hub_settler_test.go` with V2 deal mapping tests

## 7. EventMonitor V2 Event Handling

- [x] 7.1 Add `isV2Event(eventName, log)` method to detect V2 events by topic count
- [x] 7.2 Add `extractDealAndAddress(log, isV2)` helper for V1/V2 topic offset handling
- [x] 7.3 Add `extractDealID(log, isV2)` helper for deal ID extraction
- [x] 7.4 Update `handleEvent` to pass `isV2` flag and handle V2-specific event names (SettlementFinalized, EscrowOpened, MilestoneReached, DisputeRaised)
- [x] 7.5 Add `decodeAddress(log)` helper for extracting address from non-indexed data

## 8. Dangling Escrow Detector

- [x] 8.1 Create `internal/economy/escrow/hub/dangling_detector.go` with `DanglingDetector` struct
- [x] 8.2 Implement `Start`, `Stop`, `Name` lifecycle methods
- [x] 8.3 Implement `scan()` method iterating pending escrows and expiring stuck ones
- [x] 8.4 Add functional options `WithScanInterval`, `WithMaxPending`, `WithDanglingLogger`
- [x] 8.5 Publish `EscrowDanglingEvent` via event bus on detection
- [x] 8.6 Create `internal/economy/escrow/hub/dangling_detector_test.go` with unit tests

## 9. Event Bus Extension

- [x] 9.1 Add `EscrowDanglingEvent` type to `internal/eventbus/economy_events.go` with EscrowID, BuyerDID, SellerDID, Amount, PendingSince, Action fields

## 10. Config Extension

- [x] 10.1 Add `ContractVersion` field to `EscrowOnChainConfig`
- [x] 10.2 Add `HubV2Address` field to `EscrowOnChainConfig`
- [x] 10.3 Add `BeaconAddress` and `BeaconFactoryAddress` fields to `EscrowOnChainConfig`
- [x] 10.4 Add `DirectSettlerAddress` and `MilestoneSettlerAddress` fields to `EscrowOnChainConfig`
- [x] 10.5 Implement `IsV2()` method with auto-detection from V2 address presence
