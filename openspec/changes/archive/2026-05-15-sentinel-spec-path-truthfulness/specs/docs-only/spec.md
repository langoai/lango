## ADDED Requirements
### Requirement: Sentinel and economy builder-path specs stay truthful
Specs SHALL not advertise deleted app-local builder files when sentinel and economy builders live in their owning packages.

#### Scenario: Deleted builder-path claims are rejected
- **WHEN** a maintainer updates sentinel or on-chain escrow specs
- **THEN** they SHALL not claim that `internal/app/tools_sentinel.go` or `tools_economy.go` is the current source of truth
