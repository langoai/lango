## ADDED Requirements
### Requirement: CLI test harness helper-path docs stay truthful
Main specs SHALL not advertise deleted helper paths for CLI test infrastructure.

#### Scenario: Deleted harness path claims are rejected
- **WHEN** the `cli-test-harness` main spec is updated
- **THEN** it SHALL not claim that `internal/testutil/cli_harness.go` is the current shared harness implementation
