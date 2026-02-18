## 1. SQLite Connection Configuration

- [x] 1.1 Add `_journal_mode=WAL` and `_busy_timeout=5000` to the SQLite connection string in `internal/bootstrap/bootstrap.go`
- [x] 1.2 Set `db.SetMaxOpenConns(4)` and `db.SetMaxIdleConns(4)` after opening the database in bootstrap

## 2. Remove Downstream Pool Override

- [x] 2.1 Remove `db.SetMaxOpenConns(1)` from `internal/embedding/sqlite_vec.go:NewSQLiteVecStore`
- [x] 2.2 Update the function comment to reference centralized bootstrap configuration

## 3. Verification

- [x] 3.1 Run `go build ./...` to verify compilation
- [x] 3.2 Run `go test ./internal/bootstrap/ ./internal/embedding/ ./internal/knowledge/` to verify tests pass
