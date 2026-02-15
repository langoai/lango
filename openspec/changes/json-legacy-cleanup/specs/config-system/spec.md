## MODIFIED Requirements

### Requirement: Configuration save
The system SHALL save configuration through `configstore.Store.Save()` which encrypts and stores in the database. The legacy `config.Save()` function SHALL be removed.

#### Scenario: Save via configstore
- **WHEN** a config is saved through the configstore
- **THEN** it is JSON-serialized, AES-256-GCM encrypted, and stored in the database

#### Scenario: No legacy save function
- **WHEN** code attempts to call `config.Save()`
- **THEN** a compile error SHALL occur because the function no longer exists

## REMOVED Requirements

### Requirement: Configuration save (legacy)
**Reason**: The deprecated `config.Save()` function was retained for migration support from `lango.json`. With automatic JSON migration removed from bootstrap, this function is no longer needed. All configuration persistence goes through `configstore.Store.Save()`.
**Migration**: Use `configstore.Store.Save()` for encrypted profile storage.
