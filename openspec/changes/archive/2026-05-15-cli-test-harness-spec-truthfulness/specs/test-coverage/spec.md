## ADDED Requirements
### Requirement: CLI test harness spec truthfulness guard stays executable
Repository-level regressions that reintroduce deleted shared-harness path claims into the `cli-test-harness` main spec SHALL be enforced by an executable test.

#### Scenario: Deleted harness path claims are rejected
- **WHEN** the current reusable helpers live in `internal/testutil/loaders.go` and `internal/testutil/helpers.go`
- **THEN** an executable repository test SHALL fail if the `cli-test-harness` main spec claims `internal/testutil/cli_harness.go` is the current shared harness implementation
