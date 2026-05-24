## ADDED Requirements

### Requirement: Durable blocked-state cross-reference
The control-plane blocked-state surface for built-in teammate runs SHALL remain aligned with the RunLedger durability mirror defined by the `run-ledger` spec.

#### Scenario: Durable mirror does not replace live projection
- **WHEN** `agent_wait` or other live control-plane readers expose approval-blocked state
- **THEN** those readers SHALL continue using the live projection path
- **AND** the RunLedger mirror SHALL serve durable reconstruction rather than replacing the live read path in this change
