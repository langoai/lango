## ADDED Requirements

### Requirement: Confirmation depth buffer
The EventMonitor SHALL only process events up to `latest - confirmationDepth` blocks, where `confirmationDepth` is configurable and defaults to 2 for Base L2 safety.

#### Scenario: Normal processing with depth=2
- **WHEN** latest block is 100 and confirmationDepth is 2
- **THEN** EventMonitor SHALL process events up to block 98 only

#### Scenario: Zero confirmation depth (backward compatible)
- **WHEN** confirmationDepth is 0
- **THEN** EventMonitor SHALL process events up to the latest block (no buffer)

#### Scenario: Confirmation depth exceeds latest block
- **WHEN** latest block is 1 and confirmationDepth is 2
- **THEN** EventMonitor SHALL use block 1 as safeBlock (no underflow)

### Requirement: Reorg detection via block number rollback
The EventMonitor SHALL detect a reorg when `safeBlock < lastBlock` and publish an `EscrowReorgDetectedEvent` to the event bus with the previous block, new block, reorg depth, and whether the depth exceeds the confirmation buffer.

#### Scenario: Shallow reorg within confirmation depth
- **WHEN** lastBlock is 48, latest is 49, and confirmationDepth is 2 (safeBlock=47)
- **THEN** EventMonitor SHALL roll back lastBlock to 47 and publish EscrowReorgDetectedEvent with ExceedsDepth=false

#### Scenario: Deep reorg exceeding confirmation depth
- **WHEN** lastBlock is 50, latest is 47, and confirmationDepth is 2 (safeBlock=45, reorgDepth=5)
- **THEN** EventMonitor SHALL roll back lastBlock to 45 and publish EscrowReorgDetectedEvent with ExceedsDepth=true

### Requirement: Block hash cache for silent reorg detection
The EventMonitor SHALL cache block hashes and verify hash continuity on each poll cycle to detect same-height reorgs that do not change block numbers.

#### Scenario: Hash mismatch detected
- **WHEN** the cached hash for block N differs from the current hash at block N
- **THEN** EventMonitor SHALL publish EscrowReorgDetectedEvent and roll back lastBlock

#### Scenario: Cache size bounded
- **WHEN** block hash cache exceeds maxHashCache entries
- **THEN** EventMonitor SHALL trim older entries to stay within bounds

### Requirement: BlockchainClient interface
The EventMonitor SHALL depend on a `BlockchainClient` interface (HeaderByNumber, FilterLogs) instead of the concrete `*ethclient.Client` type, enabling unit testing without a real Ethereum node.

#### Scenario: Interface satisfaction
- **WHEN** `*ethclient.Client` is passed as BlockchainClient
- **THEN** compilation SHALL succeed without adapter code

### Requirement: WithConfirmationDepth option
A `WithConfirmationDepth(depth uint64)` MonitorOption SHALL be provided to configure the confirmation depth at construction time.

#### Scenario: Option applied
- **WHEN** NewEventMonitor is called with WithConfirmationDepth(5)
- **THEN** the monitor's confirmationDepth SHALL be 5
