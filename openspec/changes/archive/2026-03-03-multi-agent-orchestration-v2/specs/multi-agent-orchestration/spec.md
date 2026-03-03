## ADDED Requirements

### Requirement: Dynamic specs support in Config
The orchestration `Config` struct SHALL include a `Specs []AgentSpec` field. When non-nil, `BuildAgentTree` SHALL use these specs instead of the hardcoded built-in specs.

#### Scenario: Custom specs provided
- **WHEN** Config.Specs is set to a non-nil slice of AgentSpec
- **THEN** BuildAgentTree SHALL use those specs for agent tree construction

#### Scenario: Nil specs falls back to builtins
- **WHEN** Config.Specs is nil
- **THEN** BuildAgentTree SHALL use the default BuiltinSpecs()

### Requirement: DynamicAgents provider in Config
The orchestration `Config` struct SHALL include a `DynamicAgents` field of type `agentpool.DynamicAgentProvider`. When set, dynamic P2P agents SHALL appear in the orchestrator's routing table.

#### Scenario: P2P agents in routing table
- **WHEN** DynamicAgents is set and has available agents
- **THEN** each P2P agent SHALL appear in the routing table with "p2p:" prefix, trust score, and capabilities

#### Scenario: No P2P agents
- **WHEN** DynamicAgents is nil
- **THEN** the routing table SHALL contain only local and A2A agents

### Requirement: Capability-enhanced routing entries
Routing table entries SHALL include a `Capabilities` field listing the agent's capabilities. The orchestrator instruction SHALL display capabilities alongside agent descriptions.

#### Scenario: Routing entry with capabilities
- **WHEN** a routing entry is generated for an agent with capabilities ["search", "rag"]
- **THEN** the entry SHALL include those capabilities in the orchestrator instruction

### Requirement: DynamicToolSet and PartitionToolsDynamic
The orchestration package SHALL provide `DynamicToolSet` (map[string][]*agent.Tool) and `PartitionToolsDynamic(tools, specs)` function. The existing `PartitionTools()` SHALL be preserved as a backward-compatible wrapper.

#### Scenario: Dynamic partitioning matches static
- **WHEN** PartitionToolsDynamic is called with the built-in specs
- **THEN** the result SHALL match PartitionTools for the same tool set

#### Scenario: PartitionTools still works
- **WHEN** PartitionTools is called
- **THEN** it SHALL return the same results as before (backward compatible)
