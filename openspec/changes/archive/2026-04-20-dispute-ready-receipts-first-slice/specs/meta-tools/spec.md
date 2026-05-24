## ADDED Requirements

### Requirement: Dispute-ready receipt creation tool
The meta tools surface SHALL provide a `create_dispute_ready_receipt` tool that creates a lite submission receipt and links it to a transaction receipt.

#### Scenario: Tool creates submission and transaction linkage
- **WHEN** `create_dispute_ready_receipt` is invoked with transaction ID, artifact label, payload hash, and source lineage digest
- **THEN** it SHALL create a submission receipt
- **AND** it SHALL create or reuse the corresponding transaction receipt
- **AND** it SHALL return the created submission receipt ID, transaction receipt ID, and current submission pointer
