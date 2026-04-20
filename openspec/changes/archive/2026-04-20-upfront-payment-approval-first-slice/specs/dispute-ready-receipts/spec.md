## ADDED Requirements

### Requirement: Transaction receipt payment approval state
Transaction receipts SHALL track current payment approval state for the upfront payment path.

#### Scenario: Payment approval updates transaction state
- **WHEN** an upfront payment approval outcome is applied to a transaction receipt
- **THEN** the transaction receipt SHALL update its current payment approval status

#### Scenario: Payment approval event appended
- **WHEN** an upfront payment approval outcome is applied
- **THEN** the receipt event trail SHALL append a payment approval event for later reconstruction
