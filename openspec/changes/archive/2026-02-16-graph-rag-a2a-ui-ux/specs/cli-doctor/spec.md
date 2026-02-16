## ADDED Requirements

### Requirement: Graph store health check
The doctor command SHALL include a GraphStoreCheck that validates graph store configuration. The check SHALL skip if graph.enabled is false. When enabled, it SHALL validate that backend is "bolt", databasePath is set, and maxTraversalDepth and maxExpansionResults are positive.

#### Scenario: Graph disabled
- **WHEN** doctor runs with graph.enabled=false
- **THEN** GraphStoreCheck returns StatusSkip

#### Scenario: Graph misconfigured
- **WHEN** doctor runs with graph.enabled=true and databasePath empty
- **THEN** GraphStoreCheck returns StatusFail with message about missing path

### Requirement: Multi-agent health check
The doctor command SHALL include a MultiAgentCheck that validates multi-agent configuration. The check SHALL skip if agent.multiAgent is false. When enabled, it SHALL validate that agent.provider is set.

#### Scenario: Multi-agent disabled
- **WHEN** doctor runs with agent.multiAgent=false
- **THEN** MultiAgentCheck returns StatusSkip

### Requirement: A2A protocol health check
The doctor command SHALL include an A2ACheck that validates A2A configuration. The check SHALL skip if a2a.enabled is false. When enabled, it SHALL validate baseURL and agentName are set. Unreachable remote agents SHALL produce a warning, not a failure.

#### Scenario: A2A with unreachable remote
- **WHEN** doctor runs with a2a.enabled=true and a remote agent is unreachable
- **THEN** A2ACheck returns StatusWarn (not StatusFail)
