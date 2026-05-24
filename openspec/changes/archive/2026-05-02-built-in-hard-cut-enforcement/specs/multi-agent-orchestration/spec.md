## ADDED Requirements

### Requirement: Production ADK tree excludes built-in teammates
The production ADK agent tree SHALL keep built-in teammate types in the routing table while excluding them from `SubAgents`. Remote A2A agents and explicit non-built-in custom specs may still be attached as sub-agents.

#### Scenario: Built-in-only tree has zero production sub-agents
- **WHEN** `BuildAgentTree()` is called with only built-in teammate specs
- **THEN** the returned orchestrator SHALL expose zero built-in production sub-agents
- **AND** built-in routing information SHALL still be available to the orchestrator instruction

#### Scenario: Remote agents remain attached
- **WHEN** `BuildAgentTree()` is called with built-in teammate specs and one remote A2A agent
- **THEN** the returned orchestrator SHALL expose only the remote A2A agent as a sub-agent
