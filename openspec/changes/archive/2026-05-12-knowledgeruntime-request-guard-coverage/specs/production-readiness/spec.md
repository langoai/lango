## ADDED Requirements

### Requirement: Knowledge runtime request guards stay actionable

The knowledge-runtime service SHALL preserve actionable validation errors for missing request identifiers.

#### Scenario: Missing knowledge-runtime transaction receipt id fails closed
- **WHEN** `knowledgeruntime.Service.SelectExecutionPath` runs with an empty transaction receipt id
- **THEN** the call SHALL return an actionable `transaction_receipt_id is required` error instead of delegating to downstream store lookup behavior
