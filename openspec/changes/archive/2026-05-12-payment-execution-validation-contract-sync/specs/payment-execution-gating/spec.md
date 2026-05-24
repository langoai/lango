## MODIFIED Requirements

### Requirement: Receipt-backed direct payment execution gate
The system SHALL provide a receipt-backed direct payment execution gate for `payment_send` and `p2p_pay`. For well-formed requests, the gate SHALL return `allow` or `deny`.

#### Scenario: Missing transaction receipt id fails validation
- **WHEN** a direct payment execution request is missing `transaction_receipt_id`
- **THEN** the gate SHALL return an actionable validation error instead of a denied business result

#### Scenario: Unknown transaction receipt denies execution
- **WHEN** a direct payment execution request references a transaction receipt that does not exist
- **THEN** the gate SHALL return `deny`
