## MODIFIED Requirements

### Requirement: agent_spawn tool creates AgentRun with enriched prompt and advisory routing
The existing `agent_spawn` response shape, advisory routing semantics, and basic ID behavior SHALL remain preserved. As a refinement for the hard cut, built-in teammate production execution SHALL enter through this existing `agent_spawn` contract. `RequestedAgent` SHALL identify the built-in teammate type, and built-in production execution SHALL NOT require static ADK `transfer_to_agent` routing.

#### Scenario: Built-in teammate spawn remains the production entrypoint
- **WHEN** built-in specialist work is delegated
- **THEN** `agent_spawn` SHALL create the run
- **AND** `agent_wait` / `agent_stop` SHALL operate on that run identity chain
