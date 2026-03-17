## Why

The V1 on-chain escrow system (LangoEscrowHub) supports only simple buyer-seller deals with a single-step release model. P2P agent teams need milestone-based payments, team escrows with proportional shares, and a correlation mechanism (refId) to link on-chain deals to off-chain references. Additionally, the vault deployment model needs upgradeability via beacon proxies instead of immutable EIP-1167 clones, and the hub itself needs UUPS upgradeability for future improvements without redeployment.

## What Changes

- Add V2 Solidity contracts: `LangoEscrowHubV2` (UUPS-upgradeable hub with refId, deal types, modular settlers), `LangoVaultV2` (per-deal vault with milestone releases), `LangoBeaconVaultFactory` (beacon proxy factory for upgradeable vaults)
- Add settler strategy contracts: `DirectSettler` (immediate transfer) and `MilestoneSettler` (phased release by milestone completion)
- Add shared interfaces: `ILangoEconomy` and `ISettler` for cross-contract interoperability
- Add deployment and upgrade scripts: `DeployV2.s.sol` and `UpgradeV2.s.sol`
- Add comprehensive Foundry test suites for V2 hub (~700 lines) and V2 vault (~394 lines)
- Add Go `HubV2Client` for V2 contract interaction (refId-based deal creation, milestone management, deal type support)
- Embed V2 ABI JSON files (`LangoEscrowHubV2.abi.json`, `LangoVaultV2.abi.json`) with parsing helpers
- Add `DanglingDetector` for periodic scan of stuck pending escrows
- Extend `EventMonitor` to handle V2 events (4-topic layout with refId)
- Add V2 types (`OnChainDealV2`, `OnChainDealType`, `EscrowDanglingEvent`)
- Extend `HubSettler` with `SetDealMappingByDID` for V2 deal correlation
- Extend config with V2 fields (`HubV2Address`, `BeaconAddress`, `BeaconFactoryAddress`, `DirectSettlerAddress`, `MilestoneSettlerAddress`, `ContractVersion`, `IsV2()` helper)

## Capabilities

### New Capabilities
- `escrow-hub-v2-contracts`: V2 smart contracts with UUPS-upgradeable hub, refId-based deal correlation, modular settler strategies (Direct, Milestone), team escrow support, and beacon proxy vault factory

### Modified Capabilities
- `onchain-escrow`: V2 event handling in EventMonitor (4-topic layout detection), DanglingDetector for stuck escrow cleanup, HubSettler V2 deal mapping by DID, V2 config fields and auto-detection

## Impact

- **Contracts**: New `contracts/src/` Solidity files (hub V2, vault V2, beacon factory, settlers, interfaces) plus deployment/upgrade scripts and comprehensive test suites
- **Go packages**: `internal/economy/escrow/hub/` gains `client_v2.go`, `dangling_detector.go`, V2 ABI files, extended `abi.go`, `types.go`, `hub_settler.go`, and `monitor.go`
- **Config**: `internal/config/types_economy.go` adds 6 new fields to `EscrowOnChainConfig` plus `IsV2()` helper method
- **Event bus**: `internal/eventbus/economy_events.go` adds `EscrowDanglingEvent` type
- **Dependencies**: Requires `@openzeppelin/contracts-upgradeable` for UUPS and beacon proxy patterns
- **Backward compatibility**: Fully backward compatible; V1 contracts and clients remain unchanged; V2 is activated via config (`contractVersion: "v2"` or auto-detected from `hubV2Address`)
