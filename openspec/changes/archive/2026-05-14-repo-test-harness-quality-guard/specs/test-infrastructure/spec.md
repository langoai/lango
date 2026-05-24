## ADDED Requirements

### Requirement: Repository test harness hygiene stays reusable and deterministic
Repository test files SHALL avoid process-global stdio reassignment and legacy shared exec helpers so reusable test infrastructure remains deterministic and seam-friendly.

#### Scenario: Repository tests reject global stdio reassignment
- **WHEN** a `_test.go` file under `cmd/` or `internal/` reintroduces reassignment of `os.Stdin`, `os.Stdout`, or `os.Stderr`
- **THEN** an executable repository test SHALL fail

#### Scenario: Repository tests reject legacy ExecCmd helpers
- **WHEN** a `_test.go` file under `cmd/` or `internal/` reintroduces `testutil.ExecCmd(...)` or `testutil.ExecCmdOK(...)`
- **THEN** an executable repository test SHALL fail
