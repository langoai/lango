## ADDED Requirements

### Requirement: Session migration panic regression stays executable

Executable tests SHALL cover panic recovery in session secret migration.

#### Scenario: Panicking migration callback is covered
- **WHEN** the session package tests run
- **THEN** they SHALL exercise `MigrateSecrets` with a panicking re-encryption callback
- **AND** they SHALL assert the method returns an error instead of panicking
