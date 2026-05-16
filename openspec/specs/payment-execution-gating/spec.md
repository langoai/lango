## Purpose

Capability spec for payment-execution-gating. See requirements below for scope and behavior contracts.
## Requirements
### Requirement: Receipt-backed direct payment execution gate
The system SHALL provide a receipt-backed direct payment execution gate for `payment_send` and `p2p_pay`. For well-formed requests, the gate SHALL return `allow` or `deny`.

#### Scenario: Missing transaction receipt id fails validation
- **WHEN** a direct payment execution request is missing `transaction_receipt_id`
- **THEN** the gate SHALL return an actionable validation error instead of a denied business result

#### Scenario: Unknown transaction receipt denies execution
- **WHEN** a direct payment execution request references a transaction receipt that does not exist
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

