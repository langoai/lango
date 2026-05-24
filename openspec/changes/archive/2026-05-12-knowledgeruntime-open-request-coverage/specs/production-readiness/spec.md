## ADDED Requirements

### Requirement: Knowledge runtime open-transaction request guards stay actionable

The knowledge-runtime service SHALL preserve actionable validation errors for missing canonical open-transaction inputs.

#### Scenario: Missing canonical open inputs fail closed
- **WHEN** `knowledgeruntime.Service.OpenTransaction` runs without the required `transaction_id`, `counterparty`, or `requested_scope`
- **THEN** the call SHALL return `receipts.ErrInvalidSubmissionInput`
- **AND** SHALL preserve the actionable message that those canonical open inputs are required
