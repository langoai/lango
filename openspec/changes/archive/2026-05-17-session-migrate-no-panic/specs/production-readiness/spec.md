## ADDED Requirements

### Requirement: Session secret migration reports panics as errors

Session secret migration SHALL preserve rollback behavior and return actionable errors when migration callbacks panic instead of re-panicking into CLI callers.

#### Scenario: Re-encryption callback panic fails closed
- **WHEN** `session.EntStore.MigrateSecrets` is called and its re-encryption callback panics
- **THEN** the active transaction SHALL be rolled back
- **AND** the method SHALL return an error identifying the secret migration panic
- **AND** callers such as `lango security migrate-passphrase` SHALL receive the error through the normal migration failure path rather than a process panic
