## ADDED Requirements

### Requirement: Payment execution events in receipt trails
The receipt event trail SHALL store direct payment execution authorization and denial events.

#### Scenario: Execution authorization event appended
- **WHEN** a receipt-backed direct payment execution is allowed
- **THEN** the linked receipt trail SHALL append an authorization event

#### Scenario: Execution denial event appended
- **WHEN** a receipt-backed direct payment execution is denied
- **THEN** the linked receipt trail SHALL append a denial event with reason code
