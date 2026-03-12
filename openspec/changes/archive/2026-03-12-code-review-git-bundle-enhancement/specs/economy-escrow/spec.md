## ADDED Requirements

### Requirement: Store ListByStatusBefore filtered query
The escrow `Store` interface SHALL provide a `ListByStatusBefore(status EscrowStatus, before time.Time) []*EscrowEntry` method that returns only escrows matching the given status AND created before the specified time.

#### Scenario: Query old pending escrows
- **WHEN** `ListByStatusBefore(StatusPending, cutoffTime)` is called
- **THEN** the result SHALL contain only escrows with `Status == StatusPending` AND `CreatedAt < cutoffTime`

#### Scenario: No matching escrows
- **WHEN** `ListByStatusBefore` is called with criteria that match no entries
- **THEN** the result SHALL be an empty (or nil) slice

#### Scenario: EntStore filters at DB level
- **WHEN** the `EntStore` implementation handles `ListByStatusBefore`
- **THEN** the query SHALL use ent predicates (`escrowdeal.Status` + `escrowdeal.CreatedAtLT`) to filter at the database level rather than loading all entries into memory
