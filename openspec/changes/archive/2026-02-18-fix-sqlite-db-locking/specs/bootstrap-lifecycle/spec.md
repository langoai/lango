## MODIFIED Requirements

### Requirement: Shared database client
The bootstrap Result SHALL include the `*ent.Client` so downstream components (session store, key registry) can reuse it without opening a second connection. The underlying `*sql.DB` SHALL be configured with WAL journal mode, a busy_timeout of 5000ms, MaxOpenConns of 4, and MaxIdleConns of 4. These settings SHALL be applied in bootstrap before creating the Ent client, and no downstream component SHALL override connection pool settings on the shared `*sql.DB`.

#### Scenario: DB client reuse
- **WHEN** the bootstrap Result is passed to `app.New()`
- **THEN** the session store uses `NewEntStoreWithClient()` with the bootstrap's client

#### Scenario: WAL mode enabled at connection open
- **WHEN** the SQLite database is opened during bootstrap
- **THEN** the connection string includes `_journal_mode=WAL` and `_busy_timeout=5000`

#### Scenario: Connection pool configured centrally
- **WHEN** the `*sql.DB` is created during bootstrap
- **THEN** `MaxOpenConns` is set to 4 and `MaxIdleConns` is set to 4
- **AND** no other component overrides these settings

#### Scenario: Concurrent audit log write during active operation
- **WHEN** a background goroutine writes an audit log while another operation holds a write lock
- **THEN** the audit log write waits (up to busy_timeout) and succeeds without "database table is locked" error
