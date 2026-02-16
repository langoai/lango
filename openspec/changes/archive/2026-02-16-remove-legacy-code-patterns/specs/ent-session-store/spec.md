## REMOVED Requirements

### Requirement: SQLite backward-compatibility wrapper
**Reason**: `NewSQLiteStore` was dead code with no callers, only delegating to `NewEntStore`.
**Migration**: Use `NewEntStore` directly (already the only call path).
