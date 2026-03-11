## MODIFIED Requirements

### Requirement: On-chain escrow configuration
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

### Requirement: CLI display of confirmation depth
The `lango economy escrow show` command SHALL display the configured ConfirmationDepth value.

#### Scenario: Show command output
- **WHEN** user runs `lango economy escrow show`
- **THEN** output SHALL include "Confirmation Depth: <value>"

### Requirement: TUI settings form for confirmation depth
The on-chain escrow TUI settings form SHALL include a ConfirmationDepth input field with validation for non-negative integers.

#### Scenario: TUI form field present
- **WHEN** user opens the on-chain escrow settings form
- **THEN** a "Confirmation Depth" field SHALL be displayed with placeholder "2"
