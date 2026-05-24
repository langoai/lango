## ADDED Requirements

### Requirement: CLI test harness regressions stay command-stream based
CLI regression tests SHALL avoid process-global stdio replacement and legacy shared exec helpers after the command-stream harness migration.

#### Scenario: CLI tests reject process-global stdio replacement
- **WHEN** a CLI `_test.go` file reintroduces `os.Stdin`, `os.Stdout`, or `os.Stderr` reassignment
- **THEN** an executable repository test SHALL fail

#### Scenario: CLI tests reject legacy ExecCmd helpers
- **WHEN** a CLI `_test.go` file reintroduces `testutil.ExecCmd(...)` or `testutil.ExecCmdOK(...)`
- **THEN** an executable repository test SHALL fail
