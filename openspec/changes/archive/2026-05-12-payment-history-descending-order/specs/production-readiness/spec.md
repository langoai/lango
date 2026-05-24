## ADDED Requirements

### Requirement: Payment history ordering is directly verified

Payment history coverage SHALL verify that the newest transaction is returned first and that history limits apply on top of that descending ordering.

#### Scenario: Payment history returns newest record first
- **WHEN** payment history contains multiple records with distinct `created_at` values
- **THEN** the history response SHALL return them in descending `created_at` order

#### Scenario: Payment history limit keeps the newest record
- **WHEN** payment history is queried with `limit=1`
- **THEN** the response SHALL contain the newest record rather than an arbitrary earlier record
