## ADDED Requirements
### Requirement: CLI test harness spec does not claim deleted helper files
The `cli-test-harness` main spec SHALL describe the current reusable helpers instead of claiming a deleted shared harness file is the active implementation.

#### Scenario: Deleted harness path is not claimed
- **WHEN** a maintainer reads the `cli-test-harness` main spec
- **THEN** it SHALL not claim that `internal/testutil/cli_harness.go` is the current shared harness implementation
