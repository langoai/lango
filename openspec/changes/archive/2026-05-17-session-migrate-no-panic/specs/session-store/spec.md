## ADDED Requirements

### Requirement: Secret migration panic recovery

The session store SHALL treat recovered panics during `MigrateSecrets` as migration failures rather than rethrowing them.

#### Scenario: Panic during secret re-encryption returns error
- **WHEN** a secret re-encryption callback panics while `MigrateSecrets` is running
- **THEN** `MigrateSecrets` SHALL rollback the transaction
- **AND** it SHALL return a non-nil error
- **AND** it SHALL NOT panic
