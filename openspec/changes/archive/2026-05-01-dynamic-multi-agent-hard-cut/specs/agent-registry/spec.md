## MODIFIED Requirements

### Requirement: EmbeddedStore for default agents
The embedded built-in `AGENT.md` set SHALL remain the authoritative default inventory for built-in teammates: `operator`, `navigator`, `vault`, `librarian`, `automator`, `planner`, `chronicler`, and `ontologist`. This requirement owns which built-in defaults exist and where they come from; prompt escalation and output-handling behavior for those defaults remain governed by `agent-routing`.

#### Scenario: Load embedded defaults
- **WHEN** EmbeddedStore loads agents
- **THEN** it SHALL return 8 built-in AgentDefinitions with Source set to SourceEmbedded

#### Scenario: Embedded built-in defaults remain authoritative inventory
- **WHEN** the runtime resolves which teammate definitions are the built-in defaults
- **THEN** it SHALL use the embedded built-in inventory of `operator`, `navigator`, `vault`, `librarian`, `automator`, `planner`, `chronicler`, and `ontologist`
