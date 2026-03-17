## MODIFIED Requirements

### Requirement: Exec handlers return typed BlockedResult
The exec and exec_bg handlers SHALL return a `BlockedResult` struct instead of `map[string]interface{}` when a command is blocked. The struct SHALL have `Blocked bool` and `Message string` fields with JSON tags.

#### Scenario: Blocked command returns BlockedResult
- **WHEN** exec handler blocks a command via blockLangoExec or blockProtectedPaths
- **THEN** handler returns `&BlockedResult{Blocked: true, Message: reason}`

### Requirement: Exec handlers integrate CommandGuard
The exec and exec_bg handlers SHALL call `blockProtectedPaths` after `blockLangoExec`. The CommandGuard SHALL be constructed in `app.New()` with DataRoot and AdditionalProtectedPaths, then passed through `buildTools` → `buildExecTools`.

#### Scenario: Guard blocks protected path access
- **WHEN** agent executes `sqlite3 ~/.lango/lango.db` via exec tool
- **THEN** handler returns BlockedResult before reaching the Supervisor

#### Scenario: Guard allows normal commands
- **WHEN** agent executes `go build ./...` via exec tool
- **THEN** command passes all guards and executes normally
