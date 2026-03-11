## MODIFIED Requirements

### Requirement: DanglingDetector periodic scan
The `DanglingDetector` SHALL periodically scan for escrows stuck in `Pending` status beyond `maxPending` duration and expire them. The scan SHALL use `Store.ListByStatus(StatusPending)` instead of loading all escrows via `Store.List()`.

#### Scenario: Scan expires old pending escrows
- **WHEN** the scan runs and an escrow has been in `Pending` status longer than `maxPending`
- **THEN** the detector SHALL call `Engine.Expire` on that escrow and publish an `EscrowDanglingEvent`

#### Scenario: Scan skips non-pending escrows
- **WHEN** the scan runs
- **THEN** the detector SHALL NOT load or iterate escrows in non-pending statuses

## ADDED Requirements

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
