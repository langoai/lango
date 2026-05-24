## ADDED Requirements
### Requirement: CLI test harness spec stays truthful about helper paths
Main specs SHALL not advertise deleted shared helper paths for CLI test infrastructure.

#### Scenario: Deleted harness path claims are rejected
- **WHEN** the `cli-test-harness` main spec is updated while the shared helpers live in `internal/testutil/loaders.go` and `internal/testutil/helpers.go`
- **THEN** it SHALL not claim that `internal/testutil/cli_harness.go` is the current shared harness implementation
