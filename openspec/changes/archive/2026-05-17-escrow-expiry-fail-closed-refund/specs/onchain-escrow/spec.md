## MODIFIED Requirements

### Requirement: DanglingDetector periodic scan
The `DanglingDetector` SHALL periodically scan for escrows stuck in `Pending` status beyond `maxPending` duration and expire only those whose `ExpiresAt` has been reached. The scan SHALL use `Store.ListByStatusBefore(StatusPending, cutoff)` instead of loading all escrows via `Store.List()`.

#### Scenario: Scan expires old pending escrows
- **WHEN** the scan runs and an escrow has been in `Pending` status longer than `maxPending`
- **AND** the escrow has reached ExpiresAt
- **THEN** the detector SHALL call `Engine.Expire` on that escrow and publish an `EscrowDanglingEvent`

#### Scenario: Scan skips old pending escrows before expiry
- **WHEN** the scan runs and an escrow has been in `Pending` status longer than `maxPending`
- **AND** the escrow has not reached ExpiresAt
- **THEN** the detector SHALL leave the escrow in `Pending`
- **AND** SHALL NOT publish an `EscrowDanglingEvent`

#### Scenario: Scan skips non-pending escrows
- **WHEN** the scan runs
- **THEN** the detector SHALL NOT load or iterate escrows in non-pending statuses

#### Scenario: Configurable scan parameters
- **WHEN** `DanglingDetector` is created with `WithScanInterval(d)` and `WithMaxPending(d)` options
- **THEN** the detector SHALL scan at the specified interval and use the specified max pending threshold

#### Scenario: Lifecycle management
- **WHEN** `DanglingDetector.Start()` and `DanglingDetector.Stop()` are called
- **THEN** the detector SHALL start and stop its background scan goroutine gracefully
