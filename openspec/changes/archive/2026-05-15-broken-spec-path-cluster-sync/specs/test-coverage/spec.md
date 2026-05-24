## ADDED Requirements
### Requirement: Broken single-file path guards stay executable
Repository-level regressions that reintroduce known broken single-file references into main specs SHALL be enforced by an executable test.

#### Scenario: Known broken single-file references are rejected
- **WHEN** shared-types, skill-runtime-v2, or x402-v2 specs reintroduce `internal/cli/common/`, `cmd/main.go`, or `internal/x402/handler.go`
- **THEN** an executable repository test SHALL fail
