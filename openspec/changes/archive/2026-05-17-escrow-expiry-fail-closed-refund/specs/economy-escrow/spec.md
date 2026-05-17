## MODIFIED Requirements

### Requirement: Escrow expiration
Escrows SHALL expire after `EscrowConfig.DefaultTimeout` (default: 24h). Expired escrows transition to "expired" and funds are refunded to the buyer.

#### Scenario: Escrow expires during active work
- **WHEN** ExpiresAt is reached while escrow is "active"
- **THEN** Status transitions to "expired"
- **AND** the settlement executor SHALL refund the total amount to the buyer before the expired state is persisted

#### Scenario: Implicit expiry guard refunds locked funds
- **WHEN** a funded or active escrow has reached ExpiresAt
- **AND** a lifecycle operation such as `Activate` or `CompleteMilestone` checks expiry
- **THEN** the operation SHALL return `ErrEscrowExpired`
- **AND** the settlement executor SHALL refund the total amount to the buyer before persisting `expired`

#### Scenario: ExpiresAt boundary is expired
- **WHEN** the current time is exactly equal to ExpiresAt
- **THEN** the escrow SHALL be treated as expired
- **AND** lifecycle operations SHALL NOT proceed as if the escrow were still valid

#### Scenario: Expiry store failure is surfaced
- **WHEN** an escrow expiry path cannot persist the expired state
- **THEN** the operation SHALL return an error that includes the store update failure
- **AND** implicit expiry errors SHALL preserve `ErrEscrowExpired` in the error chain
- **AND** callers SHALL NOT receive a silent success or only a generic expiry error
- **AND** the escrow SHALL remain in its previous state when the persistence update fails

#### Scenario: Expire before timeout is rejected
- **WHEN** `Expire` is called before ExpiresAt has been reached
- **THEN** the operation SHALL return an expiry error
- **AND** the escrow SHALL remain in its previous state without refunding locked funds
