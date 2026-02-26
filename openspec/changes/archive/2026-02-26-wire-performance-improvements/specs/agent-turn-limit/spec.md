## MODIFIED Requirements

### Requirement: Configurable max turns
The agent runtime SHALL accept `maxTurns` from config via `AgentOption`. When `maxTurns` is 0 or omitted, the default (25) SHALL be used.

#### Scenario: Custom max turns from config
- **WHEN** config sets `agent.maxTurns: 15`
- **THEN** the agent SHALL enforce a 15-turn limit per run

#### Scenario: Default max turns
- **WHEN** config omits `agent.maxTurns`
- **THEN** the agent SHALL enforce a 25-turn limit per run
