## ADDED Requirements
### Requirement: Agent memory builder-path guard stays executable
Repository-level regressions that reintroduce deleted app-local agent memory builder-path claims into the `agent-memory` main spec SHALL be enforced by an executable test.

#### Scenario: Deleted agent memory builder-path claim is rejected
- **WHEN** agent memory tools are owned by `internal/agentmemory/tools.go` and wired from the current app module
- **THEN** an executable repository test SHALL fail if the `agent-memory` main spec claims `internal/app/tools_agentmemory.go` is the current source of truth
