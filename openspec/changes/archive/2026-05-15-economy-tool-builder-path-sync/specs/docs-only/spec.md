## ADDED Requirements
### Requirement: Economy tool-builder docs stay truthful about current paths
Specs SHALL not advertise deleted app-local economy or sentinel tool-builder files when those builders now live in their owning packages.

#### Scenario: Deleted builder-path claims are rejected
- **WHEN** a maintainer updates the relevant economy specs
- **THEN** they SHALL not claim that `internal/app/tools_sentinel.go` or `tools_economy.go` is the current source of truth
