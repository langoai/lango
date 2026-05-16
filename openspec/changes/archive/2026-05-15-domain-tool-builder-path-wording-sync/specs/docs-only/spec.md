## ADDED Requirements
### Requirement: Domain-tool-builders docs stay truthful about the current economy builder path
The `domain-tool-builders` main spec SHALL not describe the removed economy builder in terms of a deleted app-local file when the current source of truth is the domain-owned builder.

#### Scenario: Deleted economy builder-file wording is rejected
- **WHEN** a maintainer updates the `domain-tool-builders` main spec
- **THEN** it SHALL not describe the current economy builder ownership through `internal/app/tools_economy.go`
