## ADDED Requirements
### Requirement: CLI test harness spec reflects current helper layout
The `cli-test-harness` main spec SHALL describe the current reusable helpers in `internal/testutil/loaders.go` and `internal/testutil/helpers.go` rather than a deleted shared harness file.

#### Scenario: Deleted shared harness path is not claimed
- **WHEN** a maintainer reads the `cli-test-harness` main spec
- **THEN** it SHALL not claim that `internal/testutil/cli_harness.go` is the current shared harness implementation
