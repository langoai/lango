## Purpose

Capability spec for p2p-agent-pool. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: AgentPool management
The `p2p/agentpool` package SHALL provide a `Pool` type for managing remote P2P agents. It SHALL support Add, Get, Remove, List, and FindByCapability operations.

#### Scenario: Add and retrieve agent
- **WHEN** an agent is added to the pool
- **THEN** it SHALL be retrievable by DID via Get

#### Scenario: Find by capability
- **WHEN** FindByCapability is called with a capability string
- **THEN** it SHALL return all agents that have that capability

#### Scenario: List all agents
- **WHEN** List is called
- **THEN** it SHALL return all agents in the pool sorted by DID

### Requirement: Weighted agent scoring
The `Selector` SHALL score agents using weighted criteria: Trust (0.35), Capability (0.25), Performance (0.20), Price (0.15), Availability (0.05). The `SelectBest` method SHALL return the highest-scoring agent for a given capability.

#### Scenario: Trust dominates scoring
- **WHEN** two agents have equal capability but different trust scores
- **THEN** the agent with higher trust SHALL score higher overall

#### Scenario: SelectBest returns top agent
- **WHEN** SelectBest is called with a capability
- **THEN** it SHALL return the agent with the highest weighted score

#### Scenario: SelectN returns top N agents
- **WHEN** SelectN is called with count=3
- **THEN** it SHALL return up to 3 agents sorted by score descending

### Requirement: DynamicAgentProvider interface
The package SHALL define a `DynamicAgentProvider` interface with methods: `AvailableAgents() []DynamicAgentInfo` and `FindForCapability(capability string) []DynamicAgentInfo`.

#### Scenario: PoolProvider implements DynamicAgentProvider
- **WHEN** a PoolProvider is created with a Pool and Selector
- **THEN** it SHALL implement DynamicAgentProvider

#### Scenario: AvailableAgents returns pool contents
- **WHEN** AvailableAgents is called
- **THEN** it SHALL return info for all agents in the pool

### Requirement: HealthChecker
The package SHALL provide a `HealthChecker` that periodically pings remote agents and updates their availability status in the pool.

#### Scenario: Health check updates availability
- **WHEN** HealthChecker runs a check cycle
- **THEN** unreachable agents SHALL have their availability set to 0.0
