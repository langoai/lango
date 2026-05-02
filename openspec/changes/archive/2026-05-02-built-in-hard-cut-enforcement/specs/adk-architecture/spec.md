## ADDED Requirements

### Requirement: Hallucinated built-in targets still retry once without built-in sub-agents
The hallucinated-agent recovery path SHALL permit one correction attempt for built-in target names even when the production ADK tree has zero built-in sub-agents attached.

#### Scenario: Built-in hallucinated target retries under zero-subagent steady state
- **WHEN** a missing agent error references a built-in teammate target
- **AND** the ADK tree has zero built-in sub-agents attached
- **THEN** the runtime SHALL still emit the built-in correction hint and retry once
- **AND** it SHALL NOT suppress the correction path solely because `len(SubAgents()) == 0`
