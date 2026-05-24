## MODIFIED Requirements

### Requirement: DanglingDetector lifecycle wiring
The economy wiring SHALL create and register a DanglingDetector when on-chain escrow mode is enabled, to expire stuck pending escrows after ExpiresAt has been reached.

#### Scenario: DanglingDetector created with on-chain escrow
- **WHEN** on-chain escrow mode is enabled and an RPC client is available
- **THEN** the system SHALL create a `DanglingDetector` and register it with the lifecycle registry at `PriorityAutomation`

#### Scenario: DanglingDetector publishes events
- **WHEN** the DanglingDetector detects a stuck pending escrow that has reached ExpiresAt
- **THEN** it SHALL publish an `EscrowDanglingEvent` to the event bus
