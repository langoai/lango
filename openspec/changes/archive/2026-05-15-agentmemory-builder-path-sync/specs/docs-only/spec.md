## ADDED Requirements
### Requirement: Agent memory builder-path docs stay truthful
Main specs SHALL not advertise a deleted app-local tool-builder path for agent memory when the tools are owned by the `agentmemory` package.

#### Scenario: Deleted agent memory builder-path claim is rejected
- **WHEN** a maintainer updates the `agent-memory` main spec
- **THEN** it SHALL not claim that `internal/app/tools_agentmemory.go` is the current registration path
