## Context

The SQLite database is shared across multiple subsystems (Ent ORM, sqlite-vec embeddings, audit logging, cron, background tasks). The `SQLiteVecStore` constructor sets `SetMaxOpenConns(1)` on the shared `*sql.DB`, forcing all database operations through a single connection. Combined with the default DELETE journal mode (no WAL), any concurrent write causes immediate "database table is locked" failures.

## Goals / Non-Goals

**Goals:**
- Eliminate "database table is locked" errors during concurrent operations
- Centralize SQLite connection pool configuration in bootstrap
- Enable safe concurrent reads alongside serialized writes via WAL mode

**Non-Goals:**
- Migrating away from SQLite to another database engine
- Adding application-level retry/backoff for database operations
- Changing the fire-and-forget goroutine pattern for audit logging

## Decisions

### 1. Enable WAL journal mode via connection string pragma

WAL mode allows concurrent readers while a single writer holds the lock, and writers queue with busy_timeout instead of failing immediately.

**Alternative considered**: PRAGMA execution after connection open — rejected because connection string pragmas apply to every new connection in the pool, whereas PRAGMA execution only applies to the first connection.

### 2. Set busy_timeout=5000ms

5 seconds provides ample time for short writes (audit logs, embedding upserts) to complete without blocking indefinitely.

**Alternative considered**: Application-level retry with backoff — rejected as unnecessary complexity when SQLite's built-in busy handler achieves the same result.

### 3. Increase MaxOpenConns from 1 to 4

With WAL mode, multiple connections can read concurrently. Writes are still serialized by SQLite, but busy_timeout handles contention. 4 connections balances concurrency with SQLite's lightweight connection model.

**Alternative considered**: Unlimited connections — rejected because SQLite still serializes writes, and too many connections waste resources.

### 4. Centralize pool config in bootstrap, remove from sqlite_vec

Connection pool settings are a cross-cutting concern. Having `SQLiteVecStore` override `MaxOpenConns` on a shared `*sql.DB` is a side-effect that affects the entire application.

## Risks / Trade-offs

- [WAL mode creates `-wal` and `-shm` files alongside the database] → These are managed automatically by SQLite and cleaned up on proper close. No action needed.
- [WAL mode uses slightly more disk space] → Negligible for this application's data volume.
- [busy_timeout blocks the goroutine for up to 5s on contention] → Acceptable since write operations are short-lived and contention is transient.
