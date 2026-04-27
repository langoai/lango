## ADDED Requirements

### Requirement: Escrow release and refund tools enforce canonical adjudication
The system SHALL require release/refund executor selection to match the canonical escrow adjudication written after dispute hold.

#### Scenario: Release and refund remain fail-closed on adjudication mismatch
- **WHEN** escrow release or refund is invoked without matching adjudication
- **THEN** the tool SHALL deny execution
