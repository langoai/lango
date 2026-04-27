## ADDED Requirements

### Requirement: Receipt-backed direct payment execution gate
The system SHALL provide a receipt-backed direct payment execution gate for `payment_send` and `p2p_pay`. The gate SHALL return `allow` or `deny`.

#### Scenario: Approved prepay allows execution
- **WHEN** a direct payment execution request references a transaction receipt whose canonical payment approval state is approved and whose canonical settlement hint is `prepay`
- **THEN** the gate SHALL return `allow`

#### Scenario: Omitted submission uses current canonical submission
- **WHEN** a direct payment execution request includes `transaction_receipt_id` and omits `submission_receipt_id`
- **THEN** the gate SHALL use the transaction receipt's current canonical submission as the receipt trail target

#### Scenario: Explicit stale submission denies execution
- **WHEN** a direct payment execution request provides `submission_receipt_id` that no longer matches the transaction receipt's current canonical submission
- **THEN** the gate SHALL return `deny`

#### Scenario: Missing receipt denies execution
- **WHEN** a direct payment execution request is missing `transaction_receipt_id`
- **THEN** the gate SHALL return `deny`

#### Scenario: Non-approved payment state denies execution
- **WHEN** a direct payment execution request references a transaction receipt whose canonical payment approval state is not approved
- **THEN** the gate SHALL return `deny`

#### Scenario: Execution mode mismatch denies execution
- **WHEN** a direct payment execution request references a transaction receipt whose canonical settlement hint is not `prepay`
- **THEN** the gate SHALL return `deny`

### Requirement: Execution allow and deny evidence
The system SHALL record both authorized and denied direct payment execution outcomes into audit and receipt trails.

#### Scenario: Allowed execution recorded
- **WHEN** direct payment execution is allowed
- **THEN** the system SHALL append an authorization record to audit and receipt trail

#### Scenario: Denied execution recorded
- **WHEN** direct payment execution is denied
- **THEN** the system SHALL append a denial record with reason code to audit and receipt trail

#### Scenario: Missing evidence sink fails closed
- **WHEN** either the audit recorder or the receipt trail sink is unavailable
- **THEN** direct payment execution SHALL fail closed instead of proceeding without full evidence
