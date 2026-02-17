## MODIFIED Requirements

### Requirement: Tool partitioning for sub-agents
The `executorPrefixes` list SHALL include `"payment_"` so that payment tools are routed to the Executor sub-agent. The `capabilityMap` SHALL include `"payment_": "blockchain payments (USDC on Base)"`.

#### Scenario: Payment tools routed to executor
- **WHEN** tools are partitioned via `PartitionTools`
- **THEN** all tools with prefix `payment_` are assigned to the Executor role
