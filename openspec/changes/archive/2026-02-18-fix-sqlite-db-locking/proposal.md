## Why

SQLite concurrent write operations fail with "database table is locked" errors because the database connection pool is limited to a single connection (`SetMaxOpenConns(1)`) and WAL journal mode is not enabled. This causes background operations like audit logging to fail when competing with other write operations.

## What Changes

- Enable WAL (Write-Ahead Logging) journal mode on the SQLite connection to allow concurrent reads and serialized writes without immediate lock failures
- Set `busy_timeout=5000` so writers wait up to 5 seconds instead of failing immediately on contention
- Increase `MaxOpenConns` from 1 to 4 and set matching `MaxIdleConns` for proper connection pooling
- Remove the `SetMaxOpenConns(1)` constraint from `SQLiteVecStore` — connection pool config is now centralized in bootstrap

## Capabilities

### New Capabilities

### Modified Capabilities
- `bootstrap-lifecycle`: Connection pool and SQLite pragma configuration changes (WAL mode, busy_timeout, MaxOpenConns)

## Impact

- `internal/bootstrap/bootstrap.go` — SQLite connection string and pool settings
- `internal/embedding/sqlite_vec.go` — Removed MaxOpenConns override
- All concurrent database operations benefit (audit logging, embedding buffer, skill import, cron jobs)
