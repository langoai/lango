## ADDED Requirements
### Requirement: Sentinel and economy builder-path guard stays executable
Repository-level regressions that reintroduce deleted app-local builder-path claims into sentinel or on-chain escrow specs SHALL be enforced by an executable test.

#### Scenario: Deleted builder-path claims are rejected
- **WHEN** sentinel tools are registered by `internal/economy/escrow/sentinel/tools.go` and economy tools come from `internal/economy/tools.go`
- **THEN** an executable repository test SHALL fail if those specs claim `internal/app/tools_sentinel.go` or `tools_economy.go` is the current source of truth
