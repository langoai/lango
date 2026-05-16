## ADDED Requirements
### Requirement: CLI test harness helper-path guard stays executable
Repository-level regressions that reintroduce the deleted `internal/testutil/cli_harness.go` helper-path claim into the `cli-test-harness` main spec SHALL be enforced by an executable test.

#### Scenario: Deleted harness path claims are rejected
- **WHEN** the current reusable helpers live in `internal/testutil/loaders.go` and `internal/testutil/helpers.go`
- **THEN** an executable repository test SHALL fail if the `cli-test-harness` main spec claims `internal/testutil/cli_harness.go` is the current shared harness implementation
