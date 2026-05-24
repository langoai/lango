## ADDED Requirements

### Requirement: Escrow execution state in transaction receipts
Transaction receipts SHALL track canonical escrow execution state for escrow-backed knowledge exchange transactions.

#### Scenario: Escrow recommendation input is bound
- **WHEN** an approved upfront payment decision recommends `escrow`
- **THEN** the linked transaction receipt SHALL store escrow execution input and pending escrow execution state

#### Scenario: Escrow execution reference is preserved
- **WHEN** escrow recommendation execution creates or funds an escrow
- **THEN** the transaction receipt SHALL preserve the canonical escrow reference

### Requirement: Escrow execution events in receipt trails
The receipt event trail SHALL store escrow execution progress and failure events.

#### Scenario: Escrow progress events appended
- **WHEN** receipt-backed escrow recommendation execution starts, creates, or funds an escrow
- **THEN** the linked receipt trail SHALL append the corresponding escrow execution event

#### Scenario: Escrow failure event appended
- **WHEN** receipt-backed escrow recommendation execution fails
- **THEN** the linked receipt trail SHALL append an escrow execution failure event with failure detail
