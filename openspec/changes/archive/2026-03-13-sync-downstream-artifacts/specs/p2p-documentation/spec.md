## MODIFIED Requirements

### Requirement: P2P network documentation covers team coordination
The P2P network feature documentation SHALL include health monitoring, graceful shutdown, git state divergence detection, reorg protection, and event-driven bridges sections.

#### Scenario: Health monitoring documented
- **WHEN** a user reads `docs/features/p2p-network.md`
- **THEN** a Health Monitoring section SHALL describe periodic pings, `maxMissed` threshold, `TeamMemberUnhealthyEvent`, and config fields

#### Scenario: Graceful shutdown documented
- **WHEN** a user reads `docs/features/p2p-network.md`
- **THEN** a Graceful Shutdown section SHALL describe `TeamGracefulShutdownEvent`, shutdown sequence, and budget settlement

#### Scenario: Reorg protection documented
- **WHEN** a user reads `docs/features/p2p-network.md`
- **THEN** a Reorg Protection section SHALL describe `confirmationDepth`, `blockHashes` cache, and `EscrowReorgDetectedEvent`

#### Scenario: Event-driven bridges documented
- **WHEN** a user reads `docs/features/p2p-network.md`
- **THEN** an Event-Driven Bridges section SHALL list all 6 bridge components

### Requirement: P2P CLI documents git bundle and task branches
The P2P CLI documentation SHALL include incremental git bundle operations and task branch management commands.

#### Scenario: Incremental bundles documented
- **WHEN** a user reads `docs/cli/p2p.md`
- **THEN** `CreateIncrementalBundle`, `VerifyBundle`, `SafeApplyBundle`, `HasCommit` SHALL be documented

#### Scenario: Task branch commands documented
- **WHEN** a user reads `docs/cli/p2p.md`
- **THEN** `p2p git branch create/list/merge/delete` commands SHALL be documented with examples

### Requirement: Economy documentation covers Escrow Hub V2
The economy feature documentation SHALL include Hub V2, milestone settler, and dangling detector sections.

#### Scenario: Hub V2 documented
- **WHEN** a user reads `docs/features/economy.md`
- **THEN** a Hub V2 section SHALL describe `HubV2Client`, `DirectSettle`, milestone/team escrows, and UUPS upgradeability

#### Scenario: Dangling detector documented
- **WHEN** a user reads `docs/features/economy.md`
- **THEN** a Dangling Escrow Detector section SHALL describe `DanglingDetector`, `EscrowDanglingEvent`, and auto-refund

### Requirement: Cron documentation covers per-job timeout
The cron automation documentation SHALL include per-job timeout configuration.

#### Scenario: Per-job timeout documented
- **WHEN** a user reads `docs/automation/cron.md`
- **THEN** a Per-Job Timeout section SHALL describe `--timeout` flag, `Timeout` field, and global vs per-job precedence
