## Context

The Lango P2P economy uses on-chain escrow contracts for trustless settlement between agents. The V1 system supports simple two-party deals with a single hub contract and EIP-1167 cloned vaults. As the platform evolves to support agent teams, milestone-based deliverables, and multiple settlement strategies, the contract layer needs to be extended without breaking existing deployments.

Key constraints:
- V1 contracts are immutable once deployed; V2 must coexist alongside V1
- Go client layer must detect and support both V1 and V2 transparently
- Agent teams require proportional fund distribution to multiple members
- Milestone-based payments need on-chain enforcement, not just off-chain tracking

## Goals / Non-Goals

**Goals:**
- UUPS-upgradeable hub contract with forward-compatible storage layout
- RefId-based deal correlation between on-chain state and off-chain references (escrow IDs, team IDs)
- Modular settler architecture: pluggable settlement strategies registered on the hub
- Three deal types: Simple (direct), Milestone (phased release), Team (proportional shares)
- Beacon proxy pattern for vault deployment enabling future vault logic upgrades
- Go HubV2Client with type-safe methods for all V2 operations
- Dangling escrow detection and automatic expiration
- EventMonitor V2 support with backward-compatible V1 event handling

**Non-Goals:**
- Migrating existing V1 deals to V2 (they complete on V1)
- Cross-chain escrow (single chain only for now)
- On-chain dispute resolution AI (disputes still go to human arbitrator)
- Gas optimization beyond standard OpenZeppelin patterns

## Decisions

### 1. UUPS Upgradeability for Hub V2

**Options considered:**
- Transparent proxy (OpenZeppelin TransparentUpgradeableProxy)
- UUPS (Universal Upgradeable Proxy Standard, EIP-1822)
- Diamond pattern (EIP-2535)

**Decision:** UUPS via `@openzeppelin/contracts-upgradeable`
- UUPS puts upgrade logic in the implementation, reducing proxy contract size and gas
- Simpler admin model (no ProxyAdmin contract needed)
- Well-supported by OpenZeppelin with battle-tested libraries
- Diamond is overkill for our current contract complexity

### 2. Beacon Proxy for Vault V2

**Options considered:**
- EIP-1167 minimal proxies (used in V1)
- Beacon proxies (ERC-1967 UpgradeableBeacon)

**Decision:** Beacon proxy via `LangoBeaconVaultFactory`
- Beacon allows upgrading all vault instances simultaneously by updating the beacon implementation
- V1 EIP-1167 clones are immutable once deployed, requiring new factory deployment for fixes
- Slightly higher gas per vault creation (~20k more), but upgradability justifies it
- Factory stores beacon address; creating a vault deploys a BeaconProxy pointing to the beacon

### 3. Modular Settler Strategy Pattern

**Options considered:**
- Hard-coded settlement logic in hub contract
- Strategy pattern with registered settler contracts (ISettler interface)

**Decision:** ISettler interface + registered settlers
- Hub delegates `settle()` calls to the settler address stored per deal
- `DirectSettler`: transfers tokens immediately from hub to seller
- `MilestoneSettler`: tracks milestone completion, releases proportional amounts
- New settlers can be deployed and registered without hub upgrade
- Each deal stores its settler address at creation time

### 4. RefId-Based Deal Correlation

**Decision:** Every V2 deal includes a `bytes32 refId` parameter
- Indexed in events for efficient log filtering by off-chain ID
- Allows Go client to correlate on-chain deals with local escrow records, team IDs, or task references
- V2 events emit refId as the first indexed parameter after the event signature
- EventMonitor detects V2 events by topic count (4 topics vs V1's 3 topics)

### 5. Go V2 Client as Separate Type

**Options considered:**
- Extend existing HubClient with V2 methods
- Create separate HubV2Client type

**Decision:** Separate `HubV2Client` type
- Clean separation; V1 and V2 clients can coexist
- Different ABI JSON (hubV2ABIJSON vs hubABIJSON)
- V2 methods have different signatures (refId parameter, milestone arrays)
- Config auto-detection via `IsV2()` selects the appropriate client

### 6. Dangling Escrow Detection

**Decision:** Periodic background scanner (`DanglingDetector`)
- Polls `escrow.Store.ListByStatus(StatusPending)` at configurable interval (default: 5 min)
- Expires escrows stuck in Pending longer than `maxPending` (default: 10 min)
- Publishes `EscrowDanglingEvent` to event bus for alerting
- Implements `lifecycle.Component` for graceful start/stop

## Risks / Trade-offs

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| UUPS storage collision on upgrade | Low | Critical | Use OpenZeppelin storage gaps, comprehensive upgrade tests |
| Settler contract bug locks funds | Low | Critical | Each settler is auditable independently; emergency refund via arbitrator |
| RefId collision (bytes32 hash) | Very Low | Medium | SHA-256 of composite key provides negligible collision probability |
| V2 event detection false positive | Low | Medium | Topic count heuristic validated against known event signatures in `isV2Event()` |
| Beacon upgrade affects all vaults | Medium | High | Beacon upgrade requires owner multisig; test on staging first |
| Dangling detector false expiry | Low | Medium | Conservative `maxPending` default (10 min); logs warnings before expiring |

## Open Questions

1. ~~Should V2 hub support native ETH deals in addition to ERC-20?~~ -- No, ERC-20 only for simplicity
2. ~~Should milestone weights be enforced on-chain or off-chain?~~ -- On-chain via MilestoneSettler for trustlessness
