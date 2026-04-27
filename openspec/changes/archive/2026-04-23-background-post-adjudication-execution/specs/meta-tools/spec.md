## ADDED Requirements

### Requirement: Adjudication tool may enqueue background execution
The system SHALL allow `adjudicate_escrow_dispute` to optionally enqueue the adjudicated release or refund branch onto the background task substrate.

#### Scenario: Background dispatch returns a receipt
- **WHEN** `adjudicate_escrow_dispute` is invoked with `background_execute=true`
- **THEN** it SHALL return a background dispatch receipt
