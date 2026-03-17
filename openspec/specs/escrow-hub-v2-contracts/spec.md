## Requirements

### Requirement: UUPS-upgradeable V2 escrow hub contract
The system SHALL provide a `LangoEscrowHubV2` Solidity contract implementing UUPS upgradeability (EIP-1822) via OpenZeppelin's `UUPSUpgradeable`. The hub SHALL support three deal types (Simple, Milestone, Team) with refId-based correlation and modular settler strategies.

#### Scenario: Simple escrow creation with refId
- **WHEN** a buyer calls `createSimpleEscrow(seller, token, amount, deadline, refId)` on LangoEscrowHubV2
- **THEN** a new deal SHALL be created with type Simple, the provided refId, and the default DirectSettler, and a DealCreated event SHALL be emitted with indexed refId

#### Scenario: Milestone escrow creation
- **WHEN** a buyer calls `createMilestoneEscrow(seller, token, totalAmount, milestoneAmounts, deadline, refId)`
- **THEN** a new deal SHALL be created with type Milestone, the MilestoneSettler address, and milestone amounts stored on-chain

#### Scenario: Team escrow creation
- **WHEN** a buyer calls `createTeamEscrow(members, token, totalAmount, shares, deadline, refId)`
- **THEN** a new deal SHALL be created with type Team, proportional shares for each member, and indexed refId

#### Scenario: Direct settlement without escrow
- **WHEN** a buyer calls `directSettle(seller, token, amount, refId)`
- **THEN** tokens SHALL be transferred directly from buyer to seller without creating an escrow deal

#### Scenario: UUPS upgrade authorization
- **WHEN** a non-owner calls `upgradeTo(newImplementation)`
- **THEN** the transaction SHALL revert with an authorization error

### Requirement: V2 vault contract with milestone releases
The system SHALL provide a `LangoVaultV2` Solidity contract supporting per-deal fund custody with milestone-based partial releases. The vault SHALL be initializable for use behind beacon proxies.

#### Scenario: Vault initialization
- **WHEN** `LangoVaultV2.initialize(buyer_, seller_, token_, amount_, arbiter_, refId_)` is called (no deadline parameter — deadline is hardcoded to `block.timestamp + 30 days`)
- **THEN** the vault SHALL store deal parameters, set deadline to 30 days from initialization, and set status to Created

#### Scenario: Milestone release from vault
- **WHEN** a completed milestone triggers fund release on a vault
- **THEN** the vault SHALL transfer only the milestone's proportional amount to the seller, retaining remaining funds

### Requirement: Beacon proxy vault factory
The system SHALL provide a `LangoBeaconVaultFactory` contract that deploys vault instances as beacon proxies (ERC-1967 UpgradeableBeacon). Upgrading the beacon implementation SHALL upgrade all deployed vault proxies simultaneously.

#### Scenario: Vault creation via beacon factory
- **WHEN** `LangoBeaconVaultFactory.createVault(seller, token, amount, arbiter, refId)` is called (buyer = `msg.sender`, no deadline parameter — hardcoded to 30 days inside VaultV2)
- **THEN** a BeaconProxy clone of LangoVaultV2 SHALL be deployed and initialized with `msg.sender` as buyer, and a VaultCreated event SHALL be emitted

#### Scenario: Beacon upgrade applies to all vaults
- **WHEN** the beacon owner calls `UpgradeableBeacon.upgradeTo(newVaultImplementation)`
- **THEN** all existing vault proxies SHALL delegate to the new implementation

### Requirement: Modular settler strategy contracts
The system SHALL provide an `ISettler` interface and two settler implementations: `DirectSettler` (immediate token transfer) and `MilestoneSettler` (phased release on milestone completion). Settlers SHALL be registered on the hub and assigned per deal at creation time.

#### Scenario: ISettler.settle receives token address
- **WHEN** the hub calls `ISettler.settle()` on any settler
- **THEN** the call SHALL include `address token` as the 4th parameter: `settle(uint256 dealId, address buyer, address seller, address token, uint256 amount, bytes calldata data)`, enabling settlers to perform actual ERC-20 transfers to the seller

#### Scenario: Simple escrow uses address(0) settler
- **WHEN** a Simple escrow is created via `createSimpleEscrow`
- **THEN** the deal SHALL use `address(0)` as the settler, and the hub SHALL handle token transfers directly via `IERC20(d.token).transfer(d.seller, d.amount)` without delegating to an external settler contract

#### Scenario: DirectSettler transfers tokens to seller
- **WHEN** the hub transfers tokens to DirectSettler and calls `settle(dealId, buyer, seller, token, amount, data)`
- **THEN** DirectSettler SHALL call `IERC20(token).transfer(seller, amount)` to forward all received tokens to the seller

#### Scenario: MilestoneSettler tracks completion
- **WHEN** `MilestoneSettler.completeMilestone(dealId, index)` is called by an authorized party
- **THEN** the settler SHALL mark the milestone as completed and allow partial fund release for completed milestones only

#### Scenario: MilestoneSettler releases completed milestones
- **WHEN** the hub transfers releasable amount to MilestoneSettler and calls `settle(dealId, buyer, seller, token, releasable, data)`
- **THEN** the settler SHALL call `IERC20(token).transfer(seller, releasable)` to forward funds for all completed but unreleased milestones to the seller

### Requirement: Shared contract interfaces
The system SHALL provide `ILangoEconomy` and `ISettler` interface contracts in `contracts/src/interfaces/` for cross-contract interoperability between hub, vault, and settler contracts.

#### Scenario: Interface compliance
- **WHEN** DirectSettler and MilestoneSettler are deployed
- **THEN** both SHALL implement the `ISettler` interface

### Requirement: V2 deployment and upgrade scripts
The system SHALL provide Foundry scripts for deploying V2 contracts (`DeployV2.s.sol`) and upgrading existing proxy deployments (`UpgradeV2.s.sol`).

#### Scenario: Fresh V2 deployment
- **WHEN** `forge script DeployV2.s.sol` is executed
- **THEN** the script SHALL deploy LangoEscrowHubV2 behind a UUPS proxy, deploy settler contracts, deploy the beacon and factory, and output all contract addresses

#### Scenario: Hub V2 upgrade
- **WHEN** `forge script UpgradeV2.s.sol` is executed against an existing deployment
- **THEN** the script SHALL upgrade the hub proxy to a new implementation without losing storage state

### Requirement: Comprehensive V2 Foundry test suites
The system SHALL provide Foundry test files for LangoEscrowHubV2 (~700 lines) and LangoVaultV2 (~394 lines) covering deal lifecycle, milestone operations, dispute resolution, access control, and edge cases.

#### Scenario: Hub V2 test coverage
- **WHEN** `forge test` is run in the contracts directory
- **THEN** all LangoEscrowHubV2 tests SHALL pass covering create, deposit, release, refund, dispute, milestone, and team escrow flows

#### Scenario: Vault V2 test coverage
- **WHEN** `forge test` is run in the contracts directory
- **THEN** all LangoVaultV2 tests SHALL pass covering initialization, deposit, milestone release, and access control

### Requirement: Go HubV2Client for V2 contract interaction
The system SHALL provide a `HubV2Client` Go type in `internal/economy/escrow/hub/client_v2.go` that wraps `contract.ContractCaller` for type-safe V2 contract operations including refId-based deal creation, milestone management, and deal state queries.

#### Scenario: Create simple escrow via Go client
- **WHEN** `HubV2Client.CreateSimpleEscrow(ctx, seller, token, amount, deadline, refId)` is called
- **THEN** it SHALL call the V2 hub contract's `createSimpleEscrow` method and return the deal ID and tx hash

#### Scenario: Create milestone escrow via Go client
- **WHEN** `HubV2Client.CreateMilestoneEscrow(ctx, seller, token, totalAmount, milestoneAmounts, deadline, refId)` is called
- **THEN** it SHALL call the V2 hub contract's `createMilestoneEscrow` method and return the deal ID and tx hash

#### Scenario: Create team escrow via Go client
- **WHEN** `HubV2Client.CreateTeamEscrow(ctx, members, token, totalAmount, shares, deadline, refId)` is called
- **THEN** it SHALL call the V2 hub contract's `createTeamEscrow` method and return the deal ID and tx hash

#### Scenario: Complete and release milestones via Go client
- **WHEN** `HubV2Client.CompleteMilestone(ctx, dealID, index)` then `HubV2Client.ReleaseMilestone(ctx, dealID)` are called
- **THEN** the milestone SHALL be marked complete and funds for completed milestones SHALL be released

#### Scenario: Read V2 deal state
- **WHEN** `HubV2Client.GetDealV2(ctx, dealID)` is called
- **THEN** it SHALL return an `OnChainDealV2` struct with deal type, refId, settler address, and base deal fields

#### Scenario: Direct settlement via Go client
- **WHEN** `HubV2Client.DirectSettle(ctx, seller, token, amount, refId)` is called
- **THEN** it SHALL call the V2 hub contract's `directSettle` method and return the tx hash

### Requirement: V2 ABI embedding and parsing
The system SHALL embed `LangoEscrowHubV2.abi.json` and `LangoVaultV2.abi.json` via `//go:embed` in the hub package and provide `ParseHubV2ABI()` and `ParseVaultV2ABI()` parsing helpers.

#### Scenario: V2 ABI parsing succeeds
- **WHEN** `ParseHubV2ABI()` or `ParseVaultV2ABI()` is called
- **THEN** it SHALL return a valid `*ethabi.ABI` parsed from the embedded JSON

### Requirement: V2 Go types for deal state
The system SHALL define `OnChainDealV2` (extending `OnChainDeal` with DealType, RefId, Settler), `OnChainDealType` enum (Simple=0, Milestone=1, Team=2), and their String() methods in `internal/economy/escrow/hub/types.go`.

#### Scenario: Deal type enumeration
- **WHEN** `OnChainDealType(1).String()` is called
- **THEN** it SHALL return `"milestone"`

#### Scenario: V2 deal struct composition
- **WHEN** an `OnChainDealV2` is constructed
- **THEN** it SHALL embed `OnChainDeal` and add `DealType`, `RefId [32]byte`, and `Settler common.Address` fields

### Requirement: V2 config fields and auto-detection
The system SHALL extend `EscrowOnChainConfig` with fields `ContractVersion`, `HubV2Address`, `BeaconAddress`, `BeaconFactoryAddress`, `DirectSettlerAddress`, `MilestoneSettlerAddress` and provide an `IsV2()` method that auto-detects V2 usage from `HubV2Address` or `BeaconFactoryAddress` presence.

#### Scenario: Explicit V2 config
- **WHEN** config has `contractVersion: "v2"` set
- **THEN** `IsV2()` SHALL return true regardless of address fields

#### Scenario: Auto-detected V2 config
- **WHEN** config has `hubV2Address` set but `contractVersion` is empty
- **THEN** `IsV2()` SHALL return true

#### Scenario: V1 fallback
- **WHEN** config has `contractVersion: "v1"` set
- **THEN** `IsV2()` SHALL return false even if V2 addresses are present
