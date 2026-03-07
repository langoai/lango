## MODIFIED Requirements

### Requirement: On-chain escrow event types
The event bus SHALL support 6 additional on-chain escrow event types, each implementing `EventName() string`: EscrowOnChainDepositEvent, EscrowOnChainWorkEvent, EscrowOnChainReleaseEvent, EscrowOnChainRefundEvent, EscrowOnChainDisputeEvent, EscrowOnChainResolvedEvent. Each event SHALL include EscrowID, DealID, and TxHash fields.

#### Scenario: On-chain deposit event published
- **WHEN** EventMonitor detects a Deposited log from the hub contract
- **THEN** an EscrowOnChainDepositEvent is published with Buyer, Amount, and TxHash populated

#### Scenario: On-chain dispute event published
- **WHEN** EventMonitor detects a Disputed log from the hub contract
- **THEN** an EscrowOnChainDisputeEvent is published with Initiator and TxHash populated
