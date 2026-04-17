## MODIFIED Requirements

### Requirement: Unified bootstrap sequence
The system SHALL execute a complete bootstrap sequence with broker-owned database initialization: ensure data directory → detect encryption/header state → load envelope file → acquire credential → spawn storage broker → open database through broker → load security state/config/profile via broker-backed storage → initialize runtime services. The result SHALL be a single struct containing the initialized runtime handles, but SHALL NOT expose direct `*sql.DB` or `*ent.Client` ownership to callers once broker mode is active.

#### Scenario: Broker bootstrap on returning user
- **WHEN** bootstrap runs for a normal application start
- **THEN** the parent process SHALL spawn the storage broker before loading config profiles
- **AND** the broker SHALL own the SQLite open/migration step

#### Scenario: Broker bootstrap on first run
- **WHEN** bootstrap runs on a fresh install
- **THEN** credential acquisition and master-key setup SHALL complete before the broker `open_db` handshake is attempted
- **AND** the broker SHALL prepare the database before profile creation proceeds
