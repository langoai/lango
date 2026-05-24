## ADDED Requirements
### Requirement: Main specs stay aligned with current single-file paths
Main specs SHALL not keep stale single-file references after code moved or package paths were renamed.

#### Scenario: Known broken single-file references are rejected
- **WHEN** a maintainer updates the affected main specs
- **THEN** they SHALL not reintroduce stale references such as `internal/cli/common/`, `cmd/main.go`, or `internal/x402/handler.go`
