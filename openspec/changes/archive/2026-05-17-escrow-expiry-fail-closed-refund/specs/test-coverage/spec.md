## MODIFIED Requirements

### Requirement: Escrow lifecycle tests cover expiry settlement invariants
Escrow lifecycle tests SHALL cover both explicit and implicit expiry paths so refund and persistence failures cannot drift.

#### Scenario: Implicit expiry refund regression is covered
- **WHEN** `checkExpiry` is reached through a lifecycle method for a funded or active escrow
- **THEN** tests SHALL assert the buyer refund is invoked
- **AND** tests SHALL assert the expired state is persisted only after the refund path succeeds

#### Scenario: Implicit expiry update error is covered
- **WHEN** the store update fails during implicit expiry
- **THEN** tests SHALL assert the error is returned to the caller
- **AND** tests SHALL assert `ErrEscrowExpired` remains matchable on implicit expiry failures
- **AND** tests SHALL assert the previous escrow state is preserved

#### Scenario: Expiry boundary is covered
- **WHEN** the current time equals ExpiresAt
- **THEN** tests SHALL assert implicit and explicit expiry paths treat the escrow as expired

#### Scenario: Early explicit expiry regression is covered
- **WHEN** `Expire` is called before ExpiresAt has been reached
- **THEN** tests SHALL assert the escrow state is preserved and locked funds are not refunded

#### Scenario: Dangling detector expiry gate is covered
- **WHEN** a pending escrow is older than `maxPending` but has not reached ExpiresAt
- **THEN** tests SHALL assert the detector preserves the pending state and publishes no dangling event
- **AND** tests SHALL assert pending escrows older than `maxPending` are expired once ExpiresAt has been reached
