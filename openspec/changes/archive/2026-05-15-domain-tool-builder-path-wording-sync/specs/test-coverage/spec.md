## ADDED Requirements
### Requirement: Domain-tool-builders builder-path wording guard stays executable
Repository-level regressions that reintroduce deleted app-local economy builder wording into the `domain-tool-builders` main spec SHALL be enforced by an executable test.

#### Scenario: Deleted economy builder-file wording is rejected
- **WHEN** economy tools are built from `internal/economy/tools.go` and wired from the current app module
- **THEN** an executable repository test SHALL fail if the `domain-tool-builders` main spec describes the current source of truth through `internal/app/tools_economy.go`
