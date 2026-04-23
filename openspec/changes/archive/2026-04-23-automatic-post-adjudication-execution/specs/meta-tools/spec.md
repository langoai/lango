## ADDED Requirements

### Requirement: Adjudication tool may inline nested execution
The system SHALL allow `adjudicate_escrow_dispute` to optionally execute the matching release or refund branch inline after successful adjudication.

#### Scenario: Auto-execute uses the matching executor
- **WHEN** `adjudicate_escrow_dispute` is invoked with `auto_execute=true`
- **THEN** the matching release or refund executor SHALL be invoked inline
